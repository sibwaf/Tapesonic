package model

import (
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type LibraryArtist struct {
	Id string

	Name string

	CoverId string

	AlbumCount int

	Albums []LibraryAlbum `gorm:"-"`
}

type LibraryAlbum struct {
	Id string

	Name string

	ArtistId   string
	ArtistName string

	CoverId string

	TrackCount int
	Duration   time.Duration

	ReleasedAt *util.TimestampWrapper
	// StarredAt  *util.TimestampWrapper

	AddedAt  util.TimestampWrapper
	PlayedAt *util.TimestampWrapper
	// UpdatedAt util.TimestampWrapper

	Tracks []LibraryTrack `gorm:"-"`
}

type LibraryPlaylist struct {
	Id string

	Name string

	CoverId string

	TrackCount int
	Duration   time.Duration

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper

	Tracks []LibraryTrack `gorm:"-"`
}

type LibraryTrack struct {
	Id string

	SourceId *uuid.UUID

	RemoteId      *uuid.UUID
	RemoteTrackId *string

	Title string

	ArtistId   string
	ArtistName string

	AlbumId   string
	AlbumName string

	AlbumTrackIndex int

	CoverId string

	Duration time.Duration

	PlayedAt *util.TimestampWrapper
}

type LibraryCover struct {
	Id string

	RemoteId      uuid.UUID
	RemoteCoverId string
}
