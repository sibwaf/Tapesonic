package responses

import (
	"tapesonic/storage"
)

type TrackRs struct {
	Id       string
	SourceId string

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
		Id:            track.Id.String(),
		Artist:        track.Artist,
		Title:         track.Title,
		SourceId:      track.SourceId.String(),
		StartOffsetMs: track.StartOffsetMs,
		EndOffsetMs:   track.EndOffsetMs,
	}
}
