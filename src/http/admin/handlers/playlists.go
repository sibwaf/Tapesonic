package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/storage"
)

type playlistsHandler struct {
	playlistStorage *storage.PlaylistStorage
}

func NewPlaylistsHandler(
	playlistStorage *storage.PlaylistStorage,
) *playlistsHandler {
	return &playlistsHandler{
		playlistStorage: playlistStorage,
	}
}

func (h *playlistsHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *playlistsHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		return h.playlistStorage.GetAllPlaylists()
	case http.MethodPost:
		var playlist storage.Playlist
		err := json.NewDecoder(r.Body).Decode(&playlist)
		if err != nil {
			return nil, err
		}

		return playlist, h.playlistStorage.CreatePlaylist(&playlist)
	default:
		return nil, http.ErrNotSupported
	}
}
