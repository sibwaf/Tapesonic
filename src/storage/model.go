package storage

import "github.com/google/uuid"

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
