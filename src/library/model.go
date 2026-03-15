package library

import (
	"tapesonic/remotes"
	"tapesonic/sources"

	"github.com/google/uuid"
)

type TrackFilter struct {
	Artist string
	Title  string
	Album  string
}

type AllTrackId struct {
	Id string

	SourceTrackId *uuid.UUID
	SourceTrack   *sources.SourceTrack

	RemoteTrackId *uuid.UUID
	RemoteTrack   *remotes.RemoteTrack
}
