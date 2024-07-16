package logic

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"tapesonic/storage"
	"tapesonic/util"
	"time"
)

type StreamService struct {
	subsonic SubsonicService

	cache            *storage.StreamCacheStorage
	cacheSize        int64
	cacheMinLifetime time.Duration

	trimLock *sync.Mutex
}

func NewStreamService(
	subsonic SubsonicService,
	cache *storage.StreamCacheStorage,
	cacheSize int64,
	cacheMinLifetime time.Duration,
) *StreamService {
	return &StreamService{
		subsonic:         subsonic,
		cache:            cache,
		cacheSize:        cacheSize,
		cacheMinLifetime: cacheMinLifetime,
		trimLock:         &sync.Mutex{},
	}
}

func (svc *StreamService) Stream(id string) (mediaType string, reader io.ReadSeekCloser, err error) {
	slog.Debug(fmt.Sprintf("Using cache to stream item id=`%s`", id))

	info, reader, err := svc.cache.GetOrSave(id, func() (contentType string, reader io.ReadCloser, err error) {
		slog.Debug(fmt.Sprintf("Populating stream data cache for item id=`%s`", id))
		return svc.subsonic.Stream(context.Background(), id)
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to stream item id=`%s`: %s", id, err.Error()))
		return
	}

	slog.Debug(fmt.Sprintf("Successfully got streaming data for item id=`%s`", id))

	go func() {
		if err := svc.trimCache(); err != nil {
			slog.Error(fmt.Sprintf("Cache trimming failed: %s", err.Error()))
		}
	}()

	return info.ContentType, reader, nil
}

func (svc *StreamService) trimCache() error {
	if !svc.trimLock.TryLock() {
		return nil
	}
	defer svc.trimLock.Unlock()

	slog.Debug("Trimming stream data cache")

	for {
		cacheInfo, err := svc.cache.GetCacheInfo()
		if err != nil {
			return err
		}

		if cacheInfo.OldestItem == nil {
			slog.Debug("Stream data cache seems empty - done trimming")
			break
		}

		freeSpace := svc.cacheSize - cacheInfo.CacheSize
		spaceStatsText := fmt.Sprintf(
			"%s / %s taken, %s free",
			util.FormatBytesWithMagnitude(cacheInfo.CacheSize, svc.cacheSize),
			util.FormatBytes(svc.cacheSize),
			util.FormatBytes(freeSpace),
		)

		if freeSpace > 0 {
			slog.Debug(fmt.Sprintf("Stream data cache has enough free space - done trimming (%s)", spaceStatsText))
			break
		}

		if cacheInfo.OldestItem.AccessedAt.Add(svc.cacheMinLifetime).After(time.Now()) {
			slog.Warn(fmt.Sprintf("No suitable candidates for deletion found in stream data cache, aborting trimming (%s)", spaceStatsText))
			break
		}

		slog.Debug(
			fmt.Sprintf(
				"Deleting stream data cache item id=`%s` to free up %s (%s)",
				cacheInfo.OldestItem.Id,
				util.FormatBytes(cacheInfo.OldestItem.Size),
				spaceStatsText,
			),
		)

		err = svc.cache.Delete(cacheInfo.OldestItem.Id)
		if err != nil {
			slog.Warn(
				fmt.Sprintf(
					"Failed to delete stream data cache item id=`%s`, aborting trimming (%s): %s",
					cacheInfo.OldestItem.Id,
					spaceStatsText,
					err.Error(),
				),
			)
			break
		}
	}

	return nil
}
