package storage

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"
	"tapesonic/config"
	"tapesonic/util"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StreamCacheInfo struct {
	CacheSize  int64
	OldestItem *StreamCacheItem `gorm:"embedded"`
}

type StreamCacheItem struct {
	Id string

	Filename    string
	Size        int64
	ContentType string

	CreatedAt  util.TimestampWrapper
	AccessedAt util.TimestampWrapper
}

type StreamCacheStorage struct {
	dir         string
	maxSize     int64
	minLifetime time.Duration

	db *DbHelper

	lock     *util.StripedRwMutex
	trimLock *sync.Mutex
}

func NewStreamCacheStorage(
	dir string,
	maxSize int64,
	minLifetime time.Duration,
	db *gorm.DB,
) *StreamCacheStorage {
	return &StreamCacheStorage{
		dir:         dir,
		maxSize:     maxSize,
		minLifetime: minLifetime,

		db: NewDbHelper(db),

		lock:     util.NewStripedRwMutex(),
		trimLock: &sync.Mutex{},
	}
}

func (storage *StreamCacheStorage) GetOrSave(
	id string,
	provider func() (contentType string, reader io.ReadCloser, err error),
) (StreamCacheItem, io.ReadSeekCloser, error) {
	var itemLock *util.StripedRwMutexToken

	for {
		itemLock = storage.lock.LockForReading(id)

		item, reader, err := storage.readFile(id)
		if err == nil {
			return item, util.NewCustomCloseReadSeekCloser(reader, func() error {
				err := reader.Close()
				storage.lock.UnlockReader(id, itemLock)
				return err
			}), nil
		}

		storage.lock.UnlockReader(id, itemLock)

		// we won't get the lock if:
		//  - there's already a reader which is also trying to get the non-existing file,
		//    eventually one os us will get the write lock and fill the cache
		//    while the other one will get blocked trying to get a read lock
		//  - there's already a reader which is reading the file, we also can read
		//    the same file and there's no reason for us to use the write lock - just retry
		//  - there's a "save"-writer - just wait until it's done writing by grabbing
		//    a read lock at the start of the next iteration
		//  - there's a "delete"-writer - just wait until it's done by grabbing
		//    a read lock at the start of the next iteration

		itemLock = storage.lock.TryLockForWriting(id)
		if itemLock != nil {
			break
		}
	}

	// double-checked locking
	item, reader, err := storage.readFile(id)
	if err == nil {
		storage.lock.UnlockWriter(id, itemLock)

		// if there is a reader next in line, we'll share the file immediately
		// if there is a writer next in line, we'll get blocked until it's done
		//  - it is a "save"-writer: it has DCL logic, so it will just downgrade
		//    to a reader when it sees the same file we saw
		//  - it is a "delete"-writer: it can delete the file and reading will fail,
		// 	  but cache trimming makes sure the file wasn't accessed for a while
		//    so it's highly unlikely

		itemLock = storage.lock.LockForReading(id)

		return item, util.NewCustomCloseReadSeekCloser(reader, func() error {
			err := reader.Close()
			storage.lock.UnlockReader(id, itemLock)
			return err
		}), nil
	}

	contentType, rawReader, err := provider()
	if err != nil {
		storage.lock.UnlockWriter(id, itemLock)
		return StreamCacheItem{}, nil, err
	}

	_, err = storage.writeFile(id, contentType, rawReader)
	rawReader.Close()
	storage.lock.UnlockWriter(id, itemLock)

	if err != nil {
		return StreamCacheItem{}, nil, err
	}

	// try the full DCL again in case something happens
	// between writer unlocking and reader locking
	return storage.GetOrSave(id, provider)
}

func (storage *StreamCacheStorage) readFile(id string) (StreamCacheItem, io.ReadSeekCloser, error) {
	filename := id
	fullPath := path.Join(storage.dir, filename)

	query := `
		UPDATE stream_cache
		SET accessed_at = @accessedAt
		WHERE id = @id
		RETURNING *
	`
	params := map[string]any{
		"id":         id,
		"accessedAt": util.NewTimestampWrapper(time.Now()),
	}

	item := StreamCacheItem{}
	err := storage.db.Raw(query, params).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return StreamCacheItem{}, nil, fmt.Errorf("file with id=`%s` is not present in stream cache metadata", id)
	} else if err != nil {
		return StreamCacheItem{}, nil, err
	}

	reader, err := os.Open(fullPath)
	return item, reader, err
}

func (storage *StreamCacheStorage) writeFile(id string, contentType string, reader io.Reader) (StreamCacheItem, error) {
	filename := id
	fullPath := path.Join(storage.dir, filename)

	err := os.MkdirAll(path.Dir(fullPath), 0777)
	if err != nil {
		return StreamCacheItem{}, err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return StreamCacheItem{}, err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	size, err := io.Copy(writer, reader)
	if err != nil {
		return StreamCacheItem{}, err
	}

	sql := `
		INSERT INTO stream_cache (id, filename, size, content_type, created_at, accessed_at)
		VALUES (@id, @filename, @size, @contentType, @createdAt, @accessedAt)
		RETURNING *
	`
	params := map[string]any{
		"id":          id,
		"filename":    filename,
		"size":        size,
		"contentType": contentType,
		"createdAt":   util.NewTimestampWrapper(time.Now()),
		"accessedAt":  util.NewTimestampWrapper(time.Now()),
	}

	go func() {
		err := storage.trimToSize()
		if err != nil {
			slog.Warn(fmt.Sprintf("Stream data cache trimming failed: %s", err.Error()))
		}
	}()

	item := StreamCacheItem{}
	return item, storage.db.Raw(sql, params).First(&item).Error
}

func (storage *StreamCacheStorage) Delete(id string) error {
	itemLock := storage.lock.TryLockForWriting(id)
	if itemLock == nil {
		return errors.New("couldn't acquire a write lock")
	}
	defer storage.lock.UnlockWriter(id, itemLock)

	item := StreamCacheItem{Id: id}
	err := storage.db.Clauses(clause.Returning{}).Delete(&item).Error
	if err != nil {
		return err
	}

	if item.Filename == "" {
		return nil
	}

	fullPath := path.Join(storage.dir, item.Filename)

	err = os.Remove(fullPath)
	if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}

func (storage *StreamCacheStorage) trimToSize() error {
	if !storage.trimLock.TryLock() {
		slog.Log(context.Background(), config.LevelTrace, "Stream cache is already being trimmed, skipping")
		return nil
	}
	defer storage.trimLock.Unlock()

	slog.Debug("Trimming stream data cache")

	for {
		currentSize := int64(0)
		if err := storage.db.Raw("SELECT sum(size) FROM stream_cache").Scan(&currentSize).Error; err != nil {
			return err
		}

		spaceStatsText := fmt.Sprintf(
			"%s / %s taken, %s free",
			util.FormatBytesWithMagnitude(currentSize, storage.maxSize),
			util.FormatBytes(storage.maxSize),
			util.FormatBytes(storage.maxSize-currentSize),
		)

		if currentSize <= storage.maxSize {
			slog.Debug(fmt.Sprintf("Stream data cache has enough free space - done trimming (%s)", spaceStatsText))
			break
		}

		maxAllowedAccessedAt := time.Now().Add(-storage.minLifetime)
		nextDeletionCandidate := StreamCacheItem{}
		if err := storage.db.Where("accessed_at < ?", maxAllowedAccessedAt).Order("accessed_at ASC").Take(&nextDeletionCandidate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Warn(fmt.Sprintf("No suitable candidates for deletion found in stream data cache, aborting trimming (%s)", spaceStatsText))
				break
			} else {
				return err
			}
		}

		slog.Debug(fmt.Sprintf("Deleting stream data cache item id=`%s` to free up %s (%s)", nextDeletionCandidate.Id, util.FormatBytes(nextDeletionCandidate.Size), spaceStatsText))
		if err := storage.Delete(nextDeletionCandidate.Id); err != nil {
			return err
		}
	}

	slog.Debug("Stream data cache trimming finished")

	return nil
}
