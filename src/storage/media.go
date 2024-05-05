package storage

import (
	"path"
	"strings"

	"github.com/google/uuid"
)

type MediaStorage struct {
	dir string

	tapeStorage     *TapeStorage
	playlistStorage *PlaylistStorage
	albumStorage    *AlbumStorage
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
	albumStorage *AlbumStorage,
) *MediaStorage {
	return &MediaStorage{
		dir: dir,

		tapeStorage:     tapeStorage,
		playlistStorage: playlistStorage,
		albumStorage:    albumStorage,
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

	return CoverDescriptor{
		Path:   path.Join(ms.dir, tape.ThumbnailPath),
		Format: strings.TrimPrefix(path.Ext(tape.ThumbnailPath), "."),
	}, nil
}

func (ms *MediaStorage) GetPlaylistCover(id uuid.UUID) (CoverDescriptor, error) {
	playlist, err := ms.playlistStorage.GetPlaylistWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, playlist.ThumbnailPath),
		Format: strings.TrimPrefix(path.Ext(playlist.ThumbnailPath), "."),
	}, nil
}

func (ms *MediaStorage) GetAlbumCover(id uuid.UUID) (CoverDescriptor, error) {
	album, err := ms.albumStorage.GetAlbumWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, album.ThumbnailPath),
		Format: strings.TrimPrefix(path.Ext(album.ThumbnailPath), "."),
	}, nil
}
