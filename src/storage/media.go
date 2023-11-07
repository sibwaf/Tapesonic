package storage

import (
	"fmt"
	"path"
	"strconv"
	"strings"
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

func (ms *MediaStorage) GetTrack(id string) (TrackDescriptor, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return TrackDescriptor{}, fmt.Errorf("invalid id `%s`, expected format `tape/index`", id)
	}

	tapeId := idParts[0]
	index, err := strconv.Atoi(idParts[1])
	if err != nil {
		return TrackDescriptor{}, fmt.Errorf("invalid id `%s`, `%s` is not an index", id, idParts[1])
	}

	track, err := ms.dataStorage.GetTapeTrack(tapeId, index)
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

func (ms *MediaStorage) GetCover(id string) (CoverDescriptor, error) {
	tape, err := ms.dataStorage.GetTapeWithoutTracks(id)
	if err != nil {
		return CoverDescriptor{}, err
	}

	return CoverDescriptor{
		Path:   path.Join(ms.dir, tape.ThumbnailPath),
		Format: "png", // todo
	}, nil
}
