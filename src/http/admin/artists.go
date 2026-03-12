package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"tapesonic/artists"
	"tapesonic/model"
	"tapesonic/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ArtistListRs struct {
	Id   uuid.UUID
	Name string
}

func toArtistListRs(artist artists.Artist) ArtistListRs {
	return ArtistListRs{
		Id:   artist.Id,
		Name: artist.Name,
	}
}

func GetArtists(auth *authenticator, artistSvc *artists.ArtistService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		query := strings.TrimSpace(r.URL.Query().Get("q"))
		count := util.StringToIntOrDefault(r.URL.Query().Get("count"), 20)
		offset := util.StringToIntOrDefault(r.URL.Query().Get("offset"), 0)

		artists, err := artistSvc.Search(query, count, offset)

		return util.Map(artists, toArtistListRs), nil
	}
}

type ArtistRq struct {
	Name          string
	Aliases       []string
	MusicBrainzId string
}

func PostArtists(auth *authenticator, artists *artists.ArtistService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var rq ArtistRq
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		artist, err := artists.Create(rq.Name, rq.MusicBrainzId)
		if err != nil {
			return nil, err
		}

		return ArtistFullRs{Id: artist.Id, Name: artist.Name}, nil
	}
}

type ArtistFullRs struct {
	Id            uuid.UUID
	Name          string
	Aliases       []string
	MusicBrainzId *string
}

func toArtistFullRs(artist artists.Artist) ArtistFullRs {
	return ArtistFullRs{
		Id:            artist.Id,
		Name:          artist.Name,
		Aliases:       artist.Aliases,
		MusicBrainzId: artist.MusicBrainzId,
	}
}

func GetArtist(auth *authenticator, artists *artists.ArtistService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["artistId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		artist, err := artists.GetById(id)
		if err != nil {
			return nil, err
		}

		return toArtistFullRs(artist), nil
	}
}

func PutArtist(auth *authenticator, artists *artists.ArtistService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["artistId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		var rq ArtistRq
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		artist, err := artists.Update(id, rq.Name, rq.Aliases, rq.MusicBrainzId)
		if err != nil {
			return nil, err
		}

		return toArtistFullRs(artist), nil
	}
}

type MergeArtistsRq struct {
	Ids []uuid.UUID
}

func PostArtistsMerge(auth *authenticator, artists *artists.ArtistService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var rq MergeArtistsRq
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		result, err := artists.MergeArtists(rq.Ids)
		if err != nil {
			return nil, err
		}

		return toArtistFullRs(result), nil
	}
}
