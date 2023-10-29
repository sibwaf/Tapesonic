package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getMusicFoldersHandler struct {
}

func NewGetMusicFoldersHandler() *getMusicFoldersHandler {
	return &getMusicFoldersHandler{}
}

func (h *getMusicFoldersHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.MusicFolders = responses.NewMusicFolders([]responses.MusicFolder{})
	return response, nil
}
