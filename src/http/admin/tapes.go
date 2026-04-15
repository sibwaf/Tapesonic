package admin

import (
	"encoding/json"
	"net/http"
	"tapesonic/model"
	"tapesonic/tapes"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TapeRsArtist struct {
	Id   uuid.UUID
	Name string
}

type TapeListRs struct {
	Id   uuid.UUID
	Name string
	Type string

	ArtworkId *uuid.UUID

	Artist     *TapeRsArtist
	ReleasedAt *time.Time

	CreatedAt time.Time
}

func tapeToTapeListRs(tape tapes.SavedTape) TapeListRs {
	tapeRs := TapeListRs{
		Id:         tape.Id,
		Name:       tape.Name,
		Type:       tape.Type,
		ArtworkId:  tape.ArtworkId,
		ReleasedAt: tape.ReleasedAt.UnwrapNullable(),
		CreatedAt:  tape.CreatedAt.Unwrap(),
	}

	if tape.ArtistId != nil {
		tapeRs.Artist = &TapeRsArtist{
			Id:   *tape.ArtistId,
			Name: tape.ArtistName,
		}
	}

	return tapeRs
}

type TapeFullRs struct {
	Id   uuid.UUID
	Name string
	Type string

	ArtworkId *uuid.UUID

	Artist     *TapeRsArtist
	ReleasedAt *time.Time

	CreatedAt time.Time

	Tracks []TapeRsTrack
}

type TapeRsTrack struct {
	Id string

	SourceId *uuid.UUID
	RemoteId *uuid.UUID

	Artist *TapeRsArtist
	Title  string

	ArtworkId *uuid.UUID
}

func toTapeFullRs(tape tapes.SavedTape, tracks []model.LibraryTrack) TapeFullRs {
	tracksRs := []TapeRsTrack{}
	for _, track := range tracks {
		trackRs := TapeRsTrack{
			Id:        track.Id,
			SourceId:  track.SourceId,
			RemoteId:  track.RemoteId,
			Title:     track.Title,
			ArtworkId: track.ArtworkId,
		}

		if track.ArtistId != nil {
			trackRs.Artist = &TapeRsArtist{
				Id:   *track.ArtistId,
				Name: track.ArtistName,
			}
		}

		tracksRs = append(tracksRs, trackRs)
	}

	tapeRs := TapeFullRs{
		Id:         tape.Id,
		Name:       tape.Name,
		Type:       tape.Type,
		ArtworkId:  tape.ArtworkId,
		ReleasedAt: tape.ReleasedAt.UnwrapNullable(),
		CreatedAt:  tape.CreatedAt.Unwrap(),
		Tracks:     tracksRs,
	}

	if tape.ArtistId != nil {
		tapeRs.Artist = &TapeRsArtist{
			Id:   *tape.ArtistId,
			Name: tape.ArtistName,
		}
	}

	return tapeRs
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
	Name string
	Type string

	ArtworkId *uuid.UUID

	ArtistId   *uuid.UUID
	ReleasedAt *time.Time

	TrackIds []string
}

func tapeRqToTape(tapeRq TapeRq) tapes.Tape {
	tapeTracks := []tapes.TapeToTrack{}
	for _, trackId := range tapeRq.TrackIds {
		tapeTracks = append(tapeTracks, tapes.TapeToTrack{TrackId: trackId})
	}

	return tapes.Tape{
		Name:       tapeRq.Name,
		Type:       tapeRq.Type,
		ArtworkId:  tapeRq.ArtworkId,
		ArtistId:   tapeRq.ArtistId,
		ReleasedAt: util.NewTimestampWrapperOrNull(tapeRq.ReleasedAt),
		Tracks:     tapeTracks,
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

		savedTape, savedTracks, err := tapes.Create(user.Id, tape)
		if err != nil {
			return nil, err
		}

		return toTapeFullRs(savedTape, savedTracks), nil
	}
}

func GetTape(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["tapeId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		tape, tracks, err := tapes.GetById(user.Id, id)
		if err != nil {
			return nil, err
		}

		return toTapeFullRs(tape, tracks), nil
	}
}

func PutTape(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
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

		tape := tapeRqToTape(tapeRq)

		savedTape, savedTracks, err := tapes.Update(user.Id, id, tape)
		if err != nil {
			return nil, err
		}

		return toTapeFullRs(savedTape, savedTracks), nil
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
	TrackIds []string
}

type GuessTapeMetadataRs struct {
	Name       string
	Type       model.TapeType
	Artist     *TapeRsArtist
	ReleasedAt *time.Time
	ArtworkId  *uuid.UUID
}

func PostTapesGuessMetadata(auth *authenticator, tapes *tapes.TapeService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var guessRq GuessTapeMetadataRq
		if err := json.NewDecoder(r.Body).Decode(&guessRq); err != nil {
			return nil, err
		}

		guessedMetadata, err := tapes.GuessTapeMetadata(user.Id, guessRq.TrackIds)
		if err != nil {
			return nil, err
		}

		rs := GuessTapeMetadataRs{
			Name:       guessedMetadata.Name,
			Type:       guessedMetadata.Type,
			ReleasedAt: guessedMetadata.ReleasedAt,
			ArtworkId:  guessedMetadata.ArtworkId,
		}

		if guessedMetadata.ArtistId != nil {
			rs.Artist = &TapeRsArtist{
				Id:   *guessedMetadata.ArtistId,
				Name: guessedMetadata.ArtistName,
			}
		}

		return rs, nil
	}
}
