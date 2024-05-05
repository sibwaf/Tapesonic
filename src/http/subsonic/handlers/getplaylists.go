package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getPlaylistsHandler struct {
	subsonic logic.SubsonicService
}

func NewGetPlaylistsHandler(
	subsonic logic.SubsonicService,
) *getPlaylistsHandler {
	return &getPlaylistsHandler{
		subsonic: subsonic,
	}
}

func (h *getPlaylistsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	playlists, err := h.subsonic.GetPlaylists()
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.Playlists = playlists
	return response, nil
}
