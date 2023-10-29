package handlers

import (
	"net/http"

	"tapesonic/api/subsonic/responses"
)

type search3Handler struct {
}

func NewSearch3Handler() *search3Handler {
	return &search3Handler{}
}

func (h *search3Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.SearchResult3 = responses.NewSearchResult3(
		[]responses.ArtistId3{},
		[]responses.AlbumId3{},
		[]responses.SubsonicChild{},
	)
	return response, nil
}
