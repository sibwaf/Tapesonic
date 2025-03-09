package tasks

import (
	"fmt"
	"log/slog"
	"tapesonic/logic"
	"tapesonic/storage"
)

type DownloadSourcesTaskHandler struct {
	files   *logic.SourceFileService
	sources *storage.SourceStorage
}

func NewDownloadSourcesTaskHandler(
	files *logic.SourceFileService,
	sources *storage.SourceStorage,
) *DownloadSourcesTaskHandler {
	return &DownloadSourcesTaskHandler{
		files:   files,
		sources: sources,
	}
}

func (h *DownloadSourcesTaskHandler) Name() string {
	return "DOWNLOAD_SOURCES"
}

func (h *DownloadSourcesTaskHandler) OnSchedule() error {
	source, err := h.sources.FindNextForDownload()
	if err != nil {
		return err
	}

	if source == nil {
		slog.Debug("No sources found for download, skipping")
		return nil
	}

	slog.Debug(fmt.Sprintf("Found a source to download: id=%s, url=%s", source.Id, source.Url))

	_, err = h.files.DownloadIfMissingFor(source.Id)
	return err
}
