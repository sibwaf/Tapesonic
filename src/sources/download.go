package sources

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"tapesonic/util"
	"tapesonic/ytdlp"
	"time"

	"github.com/google/uuid"
)

type downloadService struct {
	sources *SourceStorage
	files   *FileStorage
	ytdlp   *ytdlp.YtdlpService

	baseDir string
}

func newDownloadService(
	sources *SourceStorage,
	files *FileStorage,
	ytdlp *ytdlp.YtdlpService,
	baseDir string,
) *downloadService {
	return &downloadService{
		sources: sources,
		files:   files,
		ytdlp:   ytdlp,
		baseDir: baseDir,
	}
}

func (svc *downloadService) DownloadNextPending() error {
	// todo: rework into task per source?

	source, err := svc.sources.FindNextForDownload()
	if err != nil {
		return err
	}

	if source == nil {
		return nil
	}

	slog.Debug(fmt.Sprintf("Trying to download media for source id=%s if it doesn't exist", source.Id))

	existingFile, err := svc.files.FindBySourceId(source.Id)
	if err != nil {
		return err
	}
	if existingFile != nil {
		slog.Debug(fmt.Sprintf("Source id=%s already has downloaded media (%s, %s), skipping download", source.Id, existingFile.Codec, existingFile.MediaPath))
		// todo: check that this file at least exists in the filesystem
		return nil
	}

	slog.Info(fmt.Sprintf("Downloading media for source id=%s (%s)", source.Id, source.Url))

	if source.DurationMs == 0 {
		return fmt.Errorf("source id=%s doesn't contain any media directly", source.Id)
	}

	metadata, err := svc.ytdlp.Download(source.Url, "ba", svc.baseDir)
	if err != nil {
		return err
	}

	if len(metadata.RequestedDownloads) != 1 {
		return fmt.Errorf("ytdlp returned an unexpected count=%d of downloaded files for %s", len(metadata.RequestedDownloads), source.Url)
	}

	downloadedFile := metadata.RequestedDownloads[0]
	path, err := filepath.Rel(svc.baseDir, downloadedFile.Filename)
	if err != nil {
		return fmt.Errorf("unexpected downloaded file path %s", downloadedFile.Filename)
	}

	file := SourceFile{
		Id:        uuid.New(),
		SourceId:  source.Id,
		Codec:     downloadedFile.ACodec,
		Format:    downloadedFile.Ext,
		MediaPath: path,
		CreatedAt: util.NewTimestampWrapper(time.Now()),
		UpdatedAt: util.NewTimestampWrapper(time.Now()),
	}

	slog.Info(fmt.Sprintf("Downloaded a file for source id=%s (%s): %s, %s", source.Id, source.Url, file.Codec, file.MediaPath))

	file, err = svc.files.Create(file)
	return err
}
