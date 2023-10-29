package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getArtistsHandler struct {
}

func NewGetArtistsHandler() *getArtistsHandler {
	return &getArtistsHandler{}
}

func (h *getArtistsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.Artists = responses.NewArtists("", []responses.IndexId3{})
	return response, nil
}
