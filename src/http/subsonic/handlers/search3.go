package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
	"tapesonic/util"
)

type search3Handler struct {
	subsonic logic.SubsonicService
}

func NewSearch3Handler(subsonic logic.SubsonicService) *search3Handler {
	return &search3Handler{subsonic: subsonic}
}

func (h *search3Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	searchResult3, err := h.subsonic.Search3(
		r.URL.Query().Get("query"),
		util.StringToIntOrDefault(r.URL.Query().Get("artistCount"), 20),
		util.StringToIntOrDefault(r.URL.Query().Get("artistOffset"), 0),
		util.StringToIntOrDefault(r.URL.Query().Get("albumCount"), 20),
		util.StringToIntOrDefault(r.URL.Query().Get("albumOffset"), 0),
		util.StringToIntOrDefault(r.URL.Query().Get("songCount"), 20),
		util.StringToIntOrDefault(r.URL.Query().Get("songOffset"), 0),
	)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.SearchResult3 = searchResult3
	return response, nil
}
