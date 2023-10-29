package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getRandomSongsHandler struct {
}

func NewGetRandomSongsHandler() *getRandomSongsHandler {
	return &getRandomSongsHandler{}
}

func (h *getRandomSongsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.RandomSongs = responses.NewRandomSongs([]responses.SubsonicChild{})
	return response, nil
}
