package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type albumRelatedHandler struct {
	albumStorage *storage.AlbumStorage
}

func NewAlbumRelatedHandler(
	albumStorage *storage.AlbumStorage,
) *albumRelatedHandler {
	return &albumRelatedHandler{
		albumStorage: albumStorage,
	}
}

func (h *albumRelatedHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *albumRelatedHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["albumId"]
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	return h.albumStorage.GetAlbumRelationships(id)
}
