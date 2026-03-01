package admin

import (
	"net/http"
	"tapesonic/model"
	"tapesonic/search"
	"tapesonic/util"
)

type TrackRsArtist struct {
	Id   string
	Name string
}

type TrackRs struct {
	Id       string
	SourceId string

	Artist *TrackRsArtist
	Title  string
}

func libraryToTrackRs(track model.LibraryTrack) TrackRs {
	trackRs := TrackRs{
		Id:    track.Id,
		Title: track.Title,
	}

	if track.SourceId != nil {
		trackRs.SourceId = track.SourceId.String()
	}

	if track.ArtistId != "" {
		trackRs.Artist = &TrackRsArtist{
			Id:   track.ArtistId,
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
