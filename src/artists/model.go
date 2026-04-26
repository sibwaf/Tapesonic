package artists

import "github.com/google/uuid"

// database

type Artist struct {
	Id uuid.UUID

	Name    string
	Aliases []string `gorm:"serializer:json"`

	SearchName string

	MusicBrainzId *string
}
