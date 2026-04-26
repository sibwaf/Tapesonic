package tapes

import (
	"tapesonic/model"
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

	Name      string
	Type      model.TapeType
	ArtistId  *uuid.UUID
	ArtworkId *uuid.UUID

	ReleasedAt *util.TimestampWrapper

	CreatedBy uuid.UUID
	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper

	SearchName string

	Tracks []TapeTrack
}

type TapeTrack struct {
	TapeId    uuid.UUID
	TrackId   string
	ListIndex int
}
