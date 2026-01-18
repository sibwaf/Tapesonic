package tapes

import (
	"tapesonic/model"
	"tapesonic/storage"
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
)

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
	Track   *storage.Track

	ListIndex int
}
