package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type playlistHandler struct {
	dataStorage *storage.DataStorage
}

func NewPlaylistHandler(
	dataStorage *storage.DataStorage,
) *playlistHandler {
	return &playlistHandler{
		dataStorage: dataStorage,
	}
}

func (h *playlistHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *playlistHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["playlistId"]
	id, idErr := uuid.Parse(rawId)

	switch r.Method {
	case http.MethodGet:
		if idErr != nil {
			return nil, idErr
		}
		return h.dataStorage.GetPlaylistWithTracks(id)
	default:
		return nil, http.ErrNotSupported
	}
}
