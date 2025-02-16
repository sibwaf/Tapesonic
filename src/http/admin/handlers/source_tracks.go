package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tapesonic/http/admin/requests"
	"tapesonic/http/admin/responses"
	"tapesonic/logic"
	"tapesonic/storage"
	"tapesonic/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type sourceTracksHandler struct {
	tracks *logic.TrackService
}

func NewSourceTracksHandler(
	tracks *logic.TrackService,
) *sourceTracksHandler {
	return &sourceTracksHandler{
		tracks: tracks,
	}
}

func (h *sourceTracksHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPut}
}

func (h *sourceTracksHandler) Handle(r *http.Request) (any, error) {
	sourceId, idErr := uuid.Parse(mux.Vars(r)["sourceId"])
	if idErr != nil {
		return nil, fmt.Errorf("missing or invalid sourceId")
	}

	switch r.Method {
	case http.MethodGet:
		recursive := util.StringToBoolOrDefault(r.URL.Query().Get("recursive"), false)

		var tracks []storage.Track
		var err error
		if recursive {
			tracks, err = h.tracks.GetAllTracksBySource(sourceId)
		} else {
			tracks, err = h.tracks.GetDirectTracksBySource(sourceId)
		}

		if err != nil {
			return nil, err
		}

		return responses.TracksToTrackRs(tracks), nil
	case http.MethodPut:
		var tracksRequest []requests.ModifiedTrack
		err := json.NewDecoder(r.Body).Decode(&tracksRequest)
		if err != nil {
			return nil, err
		}

		tracks := requests.ModifiedTracksToModel(tracksRequest)

		tracks, err = h.tracks.ReplaceBySource(sourceId, tracks)
		if err != nil {
			return nil, err
		}

		return responses.TracksToTrackRs(tracks), nil
	default:
		return nil, http.ErrNotSupported
	}
}
