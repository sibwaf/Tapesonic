package storage

import (
	"time"

	"github.com/google/uuid"
)

type Tape struct {
	Id uuid.UUID

	Metadata string
	Url      string

	Name       string
	AuthorName string

	ThumbnailPath string

	Tracks []*TapeTrack
}

type TapeTrack struct {
	Id uuid.UUID

	TapeId uuid.UUID
	Tape   *Tape

	FilePath string

	RawStartOffsetMs int
	StartOffsetMs    int
	RawEndOffsetMs   int
	EndOffsetMs      int

	Artist string
	Title  string

	TrackIndex int
}

type Playlist struct {
	Id uuid.UUID

	Name          string
	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	Tracks []*PlaylistTrack
}

type PlaylistTrack struct {
	Id uuid.UUID

	PlaylistId uuid.UUID
	Playlist   *Playlist

	TapeTrackId uuid.UUID
	TapeTrack   *TapeTrack

	TrackIndex int
}

type TrackDescriptor struct {
	Path          string
	StartOffsetMs int
	EndOffsetMs   int
	Format        string
}

type CoverDescriptor struct {
	Path   string
	Format string
}

type RelatedItems struct {
	Tapes []*Tape
}
