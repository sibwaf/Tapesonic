package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type albumHandler struct {
	albumStorage *storage.AlbumStorage
}

func NewAlbumHandler(
	albumStorage *storage.AlbumStorage,
) *albumHandler {
	return &albumHandler{
		albumStorage: albumStorage,
	}
}

func (h *albumHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodDelete}
}

func (h *albumHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["albumId"]
	id, idErr := uuid.Parse(rawId)

	switch r.Method {
	case http.MethodGet:
		if idErr != nil {
			return nil, idErr
		}
		return h.albumStorage.GetAlbumWithTracks(id)
	case http.MethodDelete:
		if idErr != nil {
			return nil, idErr
		}
		return nil, h.albumStorage.DeleteAlbum(id)
	default:
		return nil, http.ErrNotSupported
	}
}
