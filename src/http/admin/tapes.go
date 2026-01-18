package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tapesonic/model"
	"tapesonic/storage"
	"tapesonic/tapes"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TapeListRs struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	CreatedAt time.Time
}

func tapeToTapeListRs(tape tapes.Tape) TapeListRs {
	return TapeListRs{
		Id:   tape.Id,
		Name: tape.Name,
		Type: tape.Type,

		ThumbnailId: tape.ThumbnailId,

		Artist:     tape.Artist,
		ReleasedAt: tape.ReleasedAt.UnwrapNullable(),

		CreatedAt: tape.CreatedAt.Unwrap(),
	}
}

type TapeFullRs struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	CreatedAt time.Time

	Tracks []TapeRsTrack
}

type TapeRsTrack struct {
	Id       uuid.UUID
	SourceId uuid.UUID

	Artist string
	Title  string

	StartOffsetMs int64
	EndOffsetMs   int64
}

func tapeToTapeFullRs(tape tapes.Tape, tracks []storage.Track) TapeFullRs {
	tracksRs := []TapeRsTrack{}
	for _, track := range tracks {
		trackRs := TapeRsTrack{
			Id:            track.Id,
			Artist:        track.Artist,
			Title:         track.Title,
			SourceId:      track.SourceId,
			StartOffsetMs: track.StartOffsetMs,
			EndOffsetMs:   track.EndOffsetMs,
		}
		tracksRs = append(tracksRs, trackRs)
	}

	return TapeFullRs{
		Id:   tape.Id,
		Name: tape.Name,
		Type: tape.Type,

		ThumbnailId: tape.ThumbnailId,

		Artist:     tape.Artist,
		ReleasedAt: tape.ReleasedAt.UnwrapNullable(),

		CreatedAt: tape.CreatedAt.Unwrap(),

		Tracks: tracksRs,
	}
}

func GetTapes(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		tapes, err := tapes.GetList()
		if err != nil {
			return nil, err
		}

		return util.Map(tapes, tapeToTapeListRs), nil
	}
}

type TapeRq struct {
	Id   uuid.UUID
	Name string
	Type string

	ThumbnailId *uuid.UUID

	Artist     string
	ReleasedAt *time.Time

	Tracks []TapeRqTrack
}

type TapeRqTrack struct {
	Id uuid.UUID
}

func tapeRqToTape(tapeRq TapeRq) tapes.Tape {
	tapeToTracks := []tapes.TapeToTrack{}
	for _, track := range tapeRq.Tracks {
		tapeToTracks = append(tapeToTracks, tapes.TapeToTrack{TrackId: track.Id})
	}

	return tapes.Tape{
		Id:          tapeRq.Id,
		Name:        tapeRq.Name,
		Type:        tapeRq.Type,
		ThumbnailId: tapeRq.ThumbnailId,
		Artist:      tapeRq.Artist,
		ReleasedAt:  util.NewTimestampWrapperOrNull(tapeRq.ReleasedAt),
		Tracks:      tapeToTracks,
	}
}

func PostTapes(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var tapeRq TapeRq
		if err := json.NewDecoder(r.Body).Decode(&tapeRq); err != nil {
			return nil, err
		}

		tape := tapeRqToTape(tapeRq)

		tape, tracks, err := tapes.Create(user, tape)
		if err != nil {
			return nil, err
		}

		return tapeToTapeFullRs(tape, tracks), nil
	}
}

func GetTape(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["tapeId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		tape, tracks, err := tapes.GetById(id)
		if err != nil {
			return nil, err
		}

		return tapeToTapeFullRs(tape, tracks), nil
	}
}

func PutTape(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["tapeId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		var tapeRq TapeRq
		if err := json.NewDecoder(r.Body).Decode(&tapeRq); err != nil {
			return nil, err
		}
		if tapeRq.Id != id {
			return nil, fmt.Errorf("tapeId mismatch: tapeId=%s, tape.id=%s", id, tapeRq.Id)
		}

		tape := tapeRqToTape(tapeRq)

		tape, tracks, err := tapes.Update(tape)
		if err != nil {
			return nil, err
		}

		return tapeToTapeFullRs(tape, tracks), nil
	}
}

func DeleteTape(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["tapeId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		return nil, tapes.DeleteById(id)
	}
}

type GuessTapeMetadataRq struct {
	TrackIds []uuid.UUID
}

type GuessTapeMetadataRs struct {
	Name        string
	Type        model.TapeType
	Artist      string
	ReleasedAt  *time.Time
	ThumbnailId *uuid.UUID
}

func PostTapesGuessMetadata(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var guessRq GuessTapeMetadataRq
		if err := json.NewDecoder(r.Body).Decode(&guessRq); err != nil {
			return nil, err
		}

		guessedMetadata, err := tapes.GuessTapeMetadata(guessRq.TrackIds)
		if err != nil {
			return nil, err
		}

		return GuessTapeMetadataRs{
			Name:        guessedMetadata.Name,
			Type:        guessedMetadata.Type,
			Artist:      guessedMetadata.Artist,
			ReleasedAt:  guessedMetadata.ReleasedAt,
			ThumbnailId: guessedMetadata.ThumbnailId,
		}, nil
	}
}
