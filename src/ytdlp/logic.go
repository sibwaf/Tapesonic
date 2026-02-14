package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/semaphore"
)

type YtdlpService struct {
	ytdlp       *Ytdlp
	cache       *YtdlpMetadataStorage
	maxLifetime time.Duration
	semaphore   *semaphore.Weighted
}

func newYtdlpService(
	ytdlp *Ytdlp,
	cache *YtdlpMetadataStorage,
	maxLifetime time.Duration,
	maxParallelism int,
) *YtdlpService {
	return &YtdlpService{
		ytdlp:       ytdlp,
		cache:       cache,
		maxLifetime: maxLifetime,
		semaphore:   semaphore.NewWeighted(int64(maxParallelism)),
	}
}

func (svc *YtdlpService) GetCurrentVersion() (string, error) {
	return svc.ytdlp.GetCurrentVersion()
}

type metadataOrErr struct {
	metadata YtdlpFile
	err      error
}

func (svc *YtdlpService) GetMetadata(ctx context.Context, url string) (YtdlpFile, error) {
	cached, err := svc.cache.Find(url, time.Now().Add(-svc.maxLifetime))
	if cached != nil && err == nil {
		var metadata YtdlpFile
		err = json.Unmarshal([]byte(cached.Metadata), &metadata)
		if err == nil {
			slog.Debug(fmt.Sprintf("Returning cached metadata for %s", url))
			return metadata, nil
		}
	}

	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to get metadata for %s from cache, fetching: %s", url, err))
	} else {
		slog.Debug(fmt.Sprintf("Metadata for %s wasn't found in cache, fetching", url))
	}

	resultChannel := make(chan metadataOrErr)
	go func() {
		if err := svc.semaphore.Acquire(ctx, 1); err != nil {
			resultChannel <- metadataOrErr{err: err}
			return
		}
		defer svc.semaphore.Release(1)

		metadata, err := svc.ytdlp.ExtractMetadata(url)
		resultChannel <- metadataOrErr{metadata: metadata, err: err}
	}()

	result := <-resultChannel
	if result.err != nil {
		return YtdlpFile{}, result.err
	}

	serialized, err := json.Marshal(result.metadata)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to save metadata for %s to cache: %s", url, err))
	} else if err := svc.cache.Upsert(url, string(serialized)); err != nil {
		slog.Warn(fmt.Sprintf("Failed to save metadata for %s to cache: %s", url, err))
	}

	return result.metadata, nil
}

func (svc *YtdlpService) GetStreamInfo(ctx context.Context, url string, format string) (YtdlpFormat, error) {
	metadata, err := svc.GetMetadata(ctx, url)
	if err != nil {
		return YtdlpFormat{}, err
	}

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return YtdlpFormat{}, err
	}

	return svc.ytdlp.GetFormatFromMetadata(string(metadataStr), format)
}

func (svc *YtdlpService) Download(url string, format string, downloadDir string) (YtdlpFile, error) {
	return svc.ytdlp.Download(url, format, downloadDir)
}
