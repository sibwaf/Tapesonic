package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type scrobbleHandler struct {
}

func NewScrobbleHandler() *scrobbleHandler {
	return &scrobbleHandler{}
}

func (h *scrobbleHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	return responses.NewOkResponse(), nil
}
