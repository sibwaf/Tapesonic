package logic

import (
	"io"

	"github.com/google/uuid"
)

type AudioStream struct {
	Reader   io.ReadCloser
	MimeType string
}

type TrackProperties struct {
	SourceId uuid.UUID

	RawTitle    string
	ParentTitle string

	Artist string
	Title  string

	AlbumArtist string

	Uploader string

	StartOffsetMs int64
	EndOffsetMs   int64
}
