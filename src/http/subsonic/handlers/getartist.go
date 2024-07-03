package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getArtistHandler struct {
	subsonic logic.SubsonicService
}

func NewGetArtistHandler(
	subsonic logic.SubsonicService,
) *getArtistHandler {
	return &getArtistHandler{
		subsonic: subsonic,
	}
}

func (h *getArtistHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	artist, err := h.subsonic.GetArtist(id)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.Artist = artist
	return response, nil
}
