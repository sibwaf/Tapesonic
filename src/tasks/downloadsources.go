package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"tapesonic/config"
	"tapesonic/logic"
	"tapesonic/storage"

	"github.com/robfig/cron/v3"
)

type DownloadSourcesTaskHandler struct {
	files   *logic.SourceFileService
	sources *storage.SourceStorage

	config config.BackgroundTaskConfig
}

func NewDownloadSourcesTaskHandler(
	files *logic.SourceFileService,
	sources *storage.SourceStorage,
	config config.BackgroundTaskConfig,
) *DownloadSourcesTaskHandler {
	return &DownloadSourcesTaskHandler{
		files:   files,
		sources: sources,
		config:  config,
	}
}

func (h *DownloadSourcesTaskHandler) RegisterSchedules(cron *cron.Cron) error {
	_, err := cron.AddFunc(h.config.Cron, func() {
		err := h.trigger()
		if err != nil {
			slog.Error(fmt.Sprintf("DownloadSources task failed: %s", err.Error()))
		}
	})
	return err
}

func (h *DownloadSourcesTaskHandler) trigger() error {
	source, err := h.sources.FindNextForDownload()
	if err != nil {
		return err
	}

	if source == nil {
		slog.Log(context.Background(), config.LevelTrace, "No sources found for download, skipping")
		return nil
	}

	slog.Debug(fmt.Sprintf("Found a source to download: id=%s, url=%s", source.Id, source.Url))

	_, err = h.files.DownloadIfMissingFor(source.Id)
	return err
}
