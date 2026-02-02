package tapes

import (
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

	AlbumArtist        string
	AlbumTitle         string
	Artist             string
	SourceTitle        string
	SourceParentTitles []string `gorm:"serializer:json"`

	ReleaseDate *util.TimestampWrapper
	ThumbnailId *uuid.UUID
}

// database

type Tape struct {
	Id uuid.UUID

	Name string
	Type model.TapeType

	ThumbnailId *uuid.UUID
	Thumbnail   *storage.Thumbnail

	Tracks []TapeToTrack `gorm:"constraint:OnDelete:CASCADE;"`

	Artist     string
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
