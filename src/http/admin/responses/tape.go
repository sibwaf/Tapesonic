package responses

import (
	"tapesonic/storage"
	"time"

	"github.com/google/uuid"
)

type Tape struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	Tracks []TrackRs
}

func TapeToDto(tape storage.Tape, tracks []storage.Track) Tape {
	return Tape{
		Id:          tape.Id,
		Name:        tape.Name,
		Type:        tape.Type,
		ThumbnailId: tape.ThumbnailId,
		Artist:      tape.Artist,
		ReleasedAt:  tape.ReleasedAt,
		Tracks:      TracksToTrackRs(tracks),
	}
}

type ListTape struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	CreatedAt time.Time
}

func TapesToListDto(tapes []storage.Tape) []ListTape {
	tapeDtos := []ListTape{}
	for _, tape := range tapes {
		tapeDtos = append(tapeDtos, TapeToListDto(tape))
	}
	return tapeDtos
}

func TapeToListDto(tape storage.Tape) ListTape {
	return ListTape{
		Id:   tape.Id,
		Name: tape.Name,
		Type: tape.Type,

		ThumbnailId: tape.ThumbnailId,

		Artist:     tape.Artist,
		ReleasedAt: tape.ReleasedAt,

		CreatedAt: tape.CreatedAt,
	}
}
