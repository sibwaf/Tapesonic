package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getSongHandler struct {
	subsonic logic.SubsonicService
}

func NewGetSongHandler(
	subsonic logic.SubsonicService,
) *getSongHandler {
	return &getSongHandler{
		subsonic: subsonic,
	}
}

func (h *getSongHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	song, err := h.subsonic.GetSong(id)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.Song = song
	return response, nil
}
