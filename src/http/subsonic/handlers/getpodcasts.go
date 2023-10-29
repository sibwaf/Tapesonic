package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getPodcastsHandler struct {
}

func NewGetPodcastsHandler() *getPodcastsHandler {
	return &getPodcastsHandler{}
}

func (h *getPodcastsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.Podcasts = responses.NewPodcasts()
	return response, nil
}
