package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type playlistHandler struct {
	playlistStorage *storage.PlaylistStorage
}

func NewPlaylistHandler(
	playlistStorage *storage.PlaylistStorage,
) *playlistHandler {
	return &playlistHandler{
		playlistStorage: playlistStorage,
	}
}

func (h *playlistHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodDelete}
}

func (h *playlistHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["playlistId"]
	id, idErr := uuid.Parse(rawId)

	switch r.Method {
	case http.MethodGet:
		if idErr != nil {
			return nil, idErr
		}
		return h.playlistStorage.GetPlaylistWithTracks(id)
	case http.MethodDelete:
		if idErr != nil {
			return nil, idErr
		}
		return nil, h.playlistStorage.DeletePlaylist(id)
	default:
		return nil, http.ErrNotSupported
	}
}
