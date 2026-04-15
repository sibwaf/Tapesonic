package tapes

import (
	"tapesonic/artists"
	"tapesonic/library"
	"tapesonic/model"
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
)

// internal

type SavedTape struct {
	Id uuid.UUID

	Type      model.TapeType
	Name      string
	ArtworkId *uuid.UUID

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

	ArtworkId *uuid.UUID
	Artwork   *library.AllArtworkId

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

	TrackId string `gorm:"primaryKey"`
	Track   *library.AllTrackId

	ListIndex int
}
