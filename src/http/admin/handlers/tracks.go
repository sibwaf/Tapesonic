package handlers

import (
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"
)

type tracksHandler struct {
	tracks *logic.TrackService
	search *logic.SearchService
}

func NewTracksHandler(
	tracks *logic.TrackService,
	search *logic.SearchService,
) *tracksHandler {
	return &tracksHandler{
		tracks: tracks,
		search: search,
	}
}

func (h *tracksHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *tracksHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query().Get("q")

		result, err := h.search.SearchTracksForWebUi(q)
		if err != nil {
			return nil, err
		}

		return responses.TracksToTrackRs(result), nil
	default:
		return nil, http.ErrNotSupported
	}
}
