package artworks

import (
	"tapesonic/util"

	"github.com/google/uuid"
)

type Artwork struct {
	Id uuid.UUID

	DeduplicationId string

	FilePath string
	Format   string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type ArtworkFileDescriptor struct {
	LocalPath string
}
