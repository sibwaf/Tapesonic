package storage

import (
	"path"

	"github.com/google/uuid"
)

type MediaStorage struct {
	dir string

	tapeStorage     *TapeStorage
	playlistStorage *PlaylistStorage
}

type TrackDescriptor struct {
	Path          string
	StartOffsetMs int
	EndOffsetMs   int
	Format        string
}

type CoverDescriptor struct {
	Path   string
	Format string
}

func NewMediaStorage(
	dir string,
	tapeStorage *TapeStorage,
	playlistStorage *PlaylistStorage,
) *MediaStorage {
	return &MediaStorage{
		dir: dir,

		tapeStorage:     tapeStorage,
		playlistStorage: playlistStorage,
	}
}

func (ms *MediaStorage) GetTrack(id uuid.UUID) (TrackDescriptor, error) {
	track, err := ms.tapeStorage.GetTapeTrackWithFile(id)
	if err != nil {
		return TrackDescriptor{}, err
	}

	return TrackDescriptor{
		Path:          path.Join(ms.dir, track.TapeFile.MediaPath),
		StartOffsetMs: track.StartOffsetMs,
		EndOffsetMs:   track.EndOffsetMs,
		Format:        "opus", // todo
	}, nil
}

func (ms *MediaStorage) GetTapeCover(id uuid.UUID) (CoverDescriptor, error) {
	tape, err := ms.tapeStorage.GetTapeWithFiles(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	thumbnail := tape.ThumbnailPath
	for _, tapeFile := range tape.Files {
		if thumbnail != "" {
			break
		}

		thumbnail = tapeFile.ThumbnailPath
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, thumbnail),
		Format: "png", // todo
	}, nil
}

func (ms *MediaStorage) GetPlaylistCover(id uuid.UUID) (CoverDescriptor, error) {
	playlist, err := ms.playlistStorage.GetPlaylistWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, playlist.ThumbnailPath),
		Format: "png", // todo
	}, nil
}
