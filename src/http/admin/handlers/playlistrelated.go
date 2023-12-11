package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type playlistRelatedHandler struct {
	playlistStorage *storage.PlaylistStorage
}

func NewPlaylistRelatedHandler(
	playlistStorage *storage.PlaylistStorage,
) *playlistRelatedHandler {
	return &playlistRelatedHandler{
		playlistStorage: playlistStorage,
	}
}

func (h *playlistRelatedHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *playlistRelatedHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["playlistId"]
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	return h.playlistStorage.GetPlaylistRelationships(id)
}
