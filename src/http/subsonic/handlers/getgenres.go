package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getGenresHandler struct {
}

func NewGetGenresHandler() *getGenresHandler {
	return &getGenresHandler{}
}

func (h *getGenresHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.Genres = responses.NewGenres([]responses.Genre{})
	return response, nil
}
