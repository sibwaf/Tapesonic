package tapes

import (
	"tapesonic/artists"
	"tapesonic/model"
	"tapesonic/sources"
	"tapesonic/storage"
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
)

// internal

type TrackForMetadataGuessing struct {
	Id uuid.UUID

	AlbumArtist string
	AlbumTitle  string

	SourceTitle        string
	SourceParentTitles []string `gorm:"serializer:json"`

	ArtistId    *uuid.UUID
	ReleaseDate *util.TimestampWrapper
	ThumbnailId *uuid.UUID
}

type SavedTape struct {
	Id uuid.UUID

	Type        model.TapeType
	Name        string
	ThumbnailId *uuid.UUID

	ArtistId   *uuid.UUID
	ArtistName string

	ReleasedAt *util.TimestampWrapper

	CreatedAt util.TimestampWrapper
}

// database

type Tape struct {
	Id uuid.UUID

	Name string
	Type model.TapeType

	ThumbnailId *uuid.UUID
	Thumbnail   *storage.Thumbnail

	Tracks []TapeToTrack `gorm:"constraint:OnDelete:CASCADE;"`

	ArtistId *uuid.UUID
	Artist   *artists.Artist

	ReleasedAt *util.TimestampWrapper

	CreatedById uuid.UUID
	CreatedBy   users.User

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper

	SearchName string
}

type TapeToTrack struct {
	TapeId uuid.UUID `gorm:"primaryKey"`
	Tape   *Tape

	TrackId uuid.UUID `gorm:"primaryKey"`
	Track   *sources.SourceTrack

	ListIndex int
}
