package admin

import (
	"net/http"
	"tapesonic/model"
	"tapesonic/search"
	"tapesonic/util"

	"github.com/google/uuid"
)

type TrackRsArtist struct {
	Id   uuid.UUID
	Name string
}

type TrackRs struct {
	Id string

	SourceId *string
	RemoteId *string

	Artist *TrackRsArtist
	Title  string

	ArtworkId *uuid.UUID
}

func libraryToTrackRs(track model.LibraryTrack) TrackRs {
	trackRs := TrackRs{
		Id:        track.Id,
		Title:     track.Title,
		ArtworkId: track.ArtworkId,
	}

	if track.SourceId != nil {
		sourceId := track.SourceId.String()
		trackRs.SourceId = &sourceId
	}
	if track.RemoteId != nil {
		remoteId := track.RemoteId.String()
		trackRs.RemoteId = &remoteId
	}

	if track.ArtistId != nil {
		trackRs.Artist = &TrackRsArtist{
			Id:   *track.ArtistId,
			Name: track.ArtistName,
		}
	}

	return trackRs
}

func GetTracks(auth *authenticator, search *search.SearchService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		query := r.URL.Query().Get("q")

		tracks, err := search.FindTracksByQuery(user.Id, query)
		if err != nil {
			return nil, err
		}

		return util.Map(tracks, libraryToTrackRs), nil
	}
}
