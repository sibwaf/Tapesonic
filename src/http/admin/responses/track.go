package responses

import (
	"tapesonic/storage"

	"github.com/google/uuid"
)

type TrackRs struct {
	Id       uuid.UUID
	SourceId uuid.UUID

	Artist string
	Title  string

	StartOffsetMs int64
	EndOffsetMs   int64
}

func TracksToTrackRs(tracks []storage.Track) []TrackRs {
	trackDtos := []TrackRs{}
	for _, track := range tracks {
		trackDtos = append(trackDtos, TrackToTrackRs(track))
	}
	return trackDtos
}

func TrackToTrackRs(track storage.Track) TrackRs {
	return TrackRs{
		Id:            track.Id,
		Artist:        track.Artist,
		Title:         track.Title,
		SourceId:      track.SourceId,
		StartOffsetMs: track.StartOffsetMs,
		EndOffsetMs:   track.EndOffsetMs,
	}
}
