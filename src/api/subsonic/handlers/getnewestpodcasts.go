package handlers

import (
	"net/http"

	"tapesonic/api/subsonic/responses"
)

type getNewestPodcastsHandler struct {
}

func NewGetNewestPodcastsHandler() *getNewestPodcastsHandler {
	return &getNewestPodcastsHandler{}
}

func (h *getNewestPodcastsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.NewestPodcasts = responses.NewNewestPodcasts()
	return response, nil
}
