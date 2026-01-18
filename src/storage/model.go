package storage

import (
	"tapesonic/util"

	"github.com/google/uuid"
)

type TrackForTapeMetadataGuessing struct {
	Id uuid.UUID

	AlbumArtist        string
	AlbumTitle         string
	Artist             string
	SourceTitle        string
	SourceParentTitles []string `gorm:"serializer:json"`

	ReleaseDate *util.TimestampWrapper
	ThumbnailId *uuid.UUID
}

type SourceForHierarchy struct {
	Id       uuid.UUID
	ParentId *uuid.UUID

	Url      string
	Title    string
	Uploader string

	ListIndex int

	ThumbnailId *uuid.UUID
}
