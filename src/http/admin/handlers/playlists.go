package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/storage"
)

type playlistsHandler struct {
	dataStorage *storage.DataStorage
}

func NewPlaylistsHandler(
	dataStorage *storage.DataStorage,
) *playlistsHandler {
	return &playlistsHandler{
		dataStorage: dataStorage,
	}
}

func (h *playlistsHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *playlistsHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		return h.dataStorage.GetAllPlaylists()
	case http.MethodPost:
		var playlist storage.Playlist
		err := json.NewDecoder(r.Body).Decode(&playlist)
		if err != nil {
			return nil, err
		}

		return playlist, h.dataStorage.CreatePlaylist(&playlist)
	default:
		return nil, http.ErrNotSupported
	}
}
