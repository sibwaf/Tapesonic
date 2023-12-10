package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type playlistRelatedHandler struct {
	dataStorage *storage.DataStorage
}

func NewPlaylistRelatedHandler(
	dataStorage *storage.DataStorage,
) *playlistRelatedHandler {
	return &playlistRelatedHandler{
		dataStorage: dataStorage,
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

	return h.dataStorage.GetPlaylistRelationships(id)
}
