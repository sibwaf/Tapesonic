package library

import (
	"tapesonic/artworks"
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

type AllArtworkId struct {
	Id string

	ArtworkId *uuid.UUID
	Artwork   *artworks.Artwork

	RemoteArtworkId *uuid.UUID
	RemoteArtwork   *remotes.RemoteArtwork
}
