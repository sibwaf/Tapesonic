package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getAlbumHandler struct {
	subsonic logic.SubsonicService
}

func NewGetAlbumHandler(
	subsonic logic.SubsonicService,
) *getAlbumHandler {
	return &getAlbumHandler{
		subsonic: subsonic,
	}
}

func (h *getAlbumHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	album, err := h.subsonic.GetAlbum(id)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.Album = album
	return response, nil
}
