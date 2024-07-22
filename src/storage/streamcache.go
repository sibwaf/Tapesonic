package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
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

	CreatedAt  time.Time
	AccessedAt time.Time
}

type StreamCacheStorage struct {
	dir string
	db  *DbHelper

	lock *util.StripedRwMutex
}

func NewStreamCacheStorage(
	dir string,
	db *gorm.DB,
) (*StreamCacheStorage, error) {
	if err := db.AutoMigrate(&StreamCacheItem{}); err != nil {
		return nil, err
	}

	return &StreamCacheStorage{
		dir: dir,
		db:  NewDbHelper(db),

		lock: util.NewStripedRwMutex(),
	}, nil
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

	item := StreamCacheItem{Id: id}
	if err := storage.db.Model(&item).Clauses(clause.Returning{}).Update("accessed_at", time.Now()).Error; err != nil {
		return StreamCacheItem{}, nil, err
	}

	if item.Filename == "" {
		return StreamCacheItem{}, nil, fmt.Errorf("file with id=`%s` is not present in stream cache metadata", id)
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

	item := StreamCacheItem{
		Id:          id,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
	}

	return item, storage.db.Save(&item).Error
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

func (storage *StreamCacheStorage) GetCacheInfo() (StreamCacheInfo, error) {
	sql := `
		WITH cache_totals AS (
			SELECT sum(stream_cache_items.size) AS cache_size FROM stream_cache_items
		)
		SELECT
			stream_cache_items.*,
			cache_totals.cache_size AS cache_size
		FROM stream_cache_items, cache_totals
		ORDER BY stream_cache_items.accessed_at
		LIMIT 1
	`

	result := StreamCacheInfo{}
	return result, storage.db.Raw(sql).Find(&result).Error
}
