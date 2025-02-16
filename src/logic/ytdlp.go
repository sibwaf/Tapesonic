package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"tapesonic/storage"
	"tapesonic/ytdlp"
	"time"

	"golang.org/x/sync/semaphore"
)

type YtdlpService struct {
	ytdlp       *ytdlp.Ytdlp
	storage     *storage.YtdlpMetadataStorage
	maxLifetime time.Duration
	semaphore   *semaphore.Weighted
}

func NewYtdlpService(
	ytdlp *ytdlp.Ytdlp,
	storage *storage.YtdlpMetadataStorage,
	maxLifetime time.Duration,
	maxParallelism int,
) *YtdlpService {
	return &YtdlpService{
		ytdlp:       ytdlp,
		storage:     storage,
		maxLifetime: maxLifetime,
		semaphore:   semaphore.NewWeighted(int64(maxParallelism)),
	}
}

type metadataOrErr struct {
	metadata ytdlp.YtdlpFile
	err      error
}

func (svc *YtdlpService) GetMetadata(ctx context.Context, url string) (ytdlp.YtdlpFile, error) {
	cached, err := svc.storage.Find(url, time.Now().Add(-svc.maxLifetime))
	if cached != nil && err == nil {
		var metadata ytdlp.YtdlpFile
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
		return ytdlp.YtdlpFile{}, result.err
	}

	serialized, err := json.Marshal(result.metadata)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to save metadata for %s to cache: %s", url, err))
	} else if err := svc.storage.Upsert(url, string(serialized)); err != nil {
		slog.Warn(fmt.Sprintf("Failed to save metadata for %s to cache: %s", url, err))
	}

	return result.metadata, nil
}

func (svc *YtdlpService) GetStreamInfo(ctx context.Context, url string, format string) (ytdlp.YtdlpFormat, error) {
	metadata, err := svc.GetMetadata(ctx, url)
	if err != nil {
		return ytdlp.YtdlpFormat{}, err
	}

	metadataStr, err := json.Marshal(metadata)
	if err != nil {
		return ytdlp.YtdlpFormat{}, err
	}

	return svc.ytdlp.GetFormatFromMetadata(string(metadataStr), format)
}

func (svc *YtdlpService) Download(url string, format string, downloadDir string) (ytdlp.YtdlpFile, error) {
	return svc.ytdlp.Download(url, format, downloadDir)
}
