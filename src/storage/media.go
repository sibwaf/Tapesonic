package storage

import (
	"path"

	"github.com/google/uuid"
)

type MediaStorage struct {
	dir         string
	dataStorage *DataStorage
}

func NewMediaStorage(
	dir string,
	dataStorage *DataStorage,
) *MediaStorage {
	return &MediaStorage{
		dir:         dir,
		dataStorage: dataStorage,
	}
}

func (ms *MediaStorage) GetTrack(id uuid.UUID) (TrackDescriptor, error) {
	track, err := ms.dataStorage.GetTapeTrack(id)
	if err != nil {
		return TrackDescriptor{}, err
	}

	return TrackDescriptor{
		Path:          path.Join(ms.dir, track.FilePath),
		StartOffsetMs: track.StartOffsetMs,
		EndOffsetMs:   track.EndOffsetMs,
		Format:        "opus", // todo
	}, nil
}

func (ms *MediaStorage) GetTapeCover(id uuid.UUID) (CoverDescriptor, error) {
	tape, err := ms.dataStorage.GetTapeWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, tape.ThumbnailPath),
		Format: "png", // todo
	}, nil
}

func (ms *MediaStorage) GetPlaylistCover(id uuid.UUID) (CoverDescriptor, error) {
	playlist, err := ms.dataStorage.GetPlaylistWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, playlist.ThumbnailPath),
		Format: "png", // todo
	}, nil
}
