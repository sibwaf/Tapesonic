package requests

import (
	"tapesonic/storage"

	"github.com/google/uuid"
)

type ModifiedTrack struct {
	Id *uuid.UUID

	Artist string
	Title  string

	StartOffsetMs int64
	EndOffsetMs   int64
}

func ModifiedTracksToModel(modifiedTracks []ModifiedTrack) []storage.Track {
	tracks := []storage.Track{}
	for _, modifiedTrack := range modifiedTracks {
		tracks = append(tracks, ModifiedTrackToModel(modifiedTrack))
	}
	return tracks
}

func ModifiedTrackToModel(modifiedTrack ModifiedTrack) storage.Track {
	track := storage.Track{
		Artist:        modifiedTrack.Artist,
		Title:         modifiedTrack.Title,
		StartOffsetMs: modifiedTrack.StartOffsetMs,
		EndOffsetMs:   modifiedTrack.EndOffsetMs,
	}

	if modifiedTrack.Id != nil {
		track.Id = *modifiedTrack.Id
	}

	return track
}
