package admin

import (
	"net/http"
	"tapesonic/http/admin/responses"
	"tapesonic/model"
	"tapesonic/search"
	"tapesonic/util"
)

func libraryToTrackRs(track model.LibraryTrack) responses.TrackRs {
	sourceId := ""
	if track.SourceId != nil {
		sourceId = track.SourceId.String()
	}

	return responses.TrackRs{
		Id:       track.Id,
		SourceId: sourceId,

		Artist: track.ArtistName,
		Title:  track.Title,

		StartOffsetMs: -1,
		EndOffsetMs:   -1,
	}
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
