package requests

import (
	"tapesonic/storage"
	"time"

	"github.com/google/uuid"
)

type ModifiedTape struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	Tracks []ModifiedTapeTrack
}

type ModifiedTapeTrack struct {
	Id uuid.UUID
}

func ModifiedTapeToModel(modifiedTape ModifiedTape) storage.Tape {
	tapeToTracks := []storage.TapeToTrack{}
	for _, track := range modifiedTape.Tracks {
		tapeToTracks = append(tapeToTracks, storage.TapeToTrack{TrackId: track.Id})
	}

	return storage.Tape{
		Id:          modifiedTape.Id,
		Name:        modifiedTape.Name,
		Type:        modifiedTape.Type,
		ThumbnailId: modifiedTape.ThumbnailId,
		Artist:      modifiedTape.Artist,
		ReleasedAt:  modifiedTape.ReleasedAt,
		Tracks:      tapeToTracks,
	}
}
