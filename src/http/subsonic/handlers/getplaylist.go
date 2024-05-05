package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getPlaylistHandler struct {
	subsonic logic.SubsonicService
}

func NewGetPlaylistHandler(
	subsonic logic.SubsonicService,
) *getPlaylistHandler {
	return &getPlaylistHandler{
		subsonic: subsonic,
	}
}

func (h *getPlaylistHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	playlist, err := h.subsonic.GetPlaylist(id)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.Playlist = playlist
	return response, nil
}
