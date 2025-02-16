package logic

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type SourceFileService struct {
	storage *storage.SourceFileStorage
	sources *storage.SourceStorage
	ytdlp   *YtdlpService

	dir string
}

func NewSourceFileService(
	storage *storage.SourceFileStorage,
	sources *storage.SourceStorage,
	ytdlp *YtdlpService,
	dir string,
) *SourceFileService {
	return &SourceFileService{
		storage: storage,
		sources: sources,
		ytdlp:   ytdlp,
		dir:     dir,
	}
}

func (s *SourceFileService) DeleteFor(sourceId uuid.UUID) error {
	slog.Debug(fmt.Sprintf("Trying to delete media for source id=%s", sourceId))

	file, err := s.storage.FindBySourceId(sourceId)
	if err != nil {
		return err
	}

	if file == nil {
		slog.Debug(fmt.Sprintf("No file metadata found for source id=%s, nothing to delete", sourceId))
		return nil
	}

	mediaPath := path.Join(s.dir, file.MediaPath)
	slog.Debug(fmt.Sprintf("Deleting file id=%s (%s) for source id=%s", file.Id, mediaPath, sourceId))

	err = os.Remove(file.MediaPath)
	if errors.Is(err, os.ErrNotExist) {
		slog.Debug(fmt.Sprintf("File id=%s (%s) for source id=%s doesn't exist in FS, deleting metadata", file.Id, mediaPath, sourceId))
	} else if err != nil {
		return err
	}

	err = s.storage.DeleteById(file.Id)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Deleted file id=%s (%s) for source id=%s", file.Id, mediaPath, sourceId))
	return nil
}

func (s *SourceFileService) DownloadIfMissingFor(sourceId uuid.UUID) (storage.SourceFile, error) {
	slog.Debug(fmt.Sprintf("Trying to download media for source id=%s if it doesn't exist", sourceId))

	existingFile, err := s.storage.FindBySourceId(sourceId)
	if err != nil {
		return storage.SourceFile{}, err
	}
	if existingFile != nil {
		slog.Debug(fmt.Sprintf("Source id=%s already has downloaded media (%s, %s), skipping download", sourceId, existingFile.Codec, existingFile.MediaPath))
		// todo: check that this file at least exists in the filesystem
		return *existingFile, nil
	}

	source, err := s.sources.GetById(sourceId)
	if err != nil {
		return storage.SourceFile{}, err
	}

	slog.Info(fmt.Sprintf("Downloading media for source id=%s (%s)", source.Id, source.Url))

	if source.DurationMs == 0 {
		return storage.SourceFile{}, fmt.Errorf("source id=%s doesn't contain any media directly", sourceId)
	}

	metadata, err := s.ytdlp.Download(source.Url, "ba", s.dir)
	if err != nil {
		return storage.SourceFile{}, err
	}

	if len(metadata.RequestedDownloads) != 1 {
		return storage.SourceFile{}, fmt.Errorf("ytdlp returned an unexpected count=%d of downloaded files for %s", len(metadata.RequestedDownloads), source.Url)
	}

	downloadedFile := metadata.RequestedDownloads[0]
	path, err := filepath.Rel(s.dir, downloadedFile.Filename)
	if err != nil {
		return storage.SourceFile{}, fmt.Errorf("unexpected downloaded file path %s", downloadedFile.Filename)
	}

	file := storage.SourceFile{
		SourceId:  source.Id,
		Codec:     downloadedFile.ACodec,
		Format:    downloadedFile.Ext,
		MediaPath: path,
	}

	slog.Info(fmt.Sprintf("Downloaded a file for source id=%s (%s): %s, %s", source.Id, source.Url, file.Codec, file.MediaPath))

	return s.storage.Create(file)
}

func (s *SourceFileService) FindBySourceId(sourceId uuid.UUID) (*storage.SourceFile, error) {
	return s.storage.FindBySourceId(sourceId)
}

func (s *SourceFileService) FindBySourceIds(sourceIds []uuid.UUID) ([]storage.SourceFile, error) {
	return s.storage.FindBySourceIds(sourceIds)
}
