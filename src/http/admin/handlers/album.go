package handlers

import (
	"encoding/json"
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
	return []string{http.MethodGet, http.MethodPut, http.MethodDelete}
}

func (h *albumHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["albumId"]
	id, idErr := uuid.Parse(rawId)
	if idErr != nil {
		return nil, idErr
	}

	switch r.Method {
	case http.MethodGet:
		return h.albumStorage.GetAlbumWithTracks(id)
	case http.MethodPut:
		var album storage.Album
		err := json.NewDecoder(r.Body).Decode(&album)
		if err != nil {
			return nil, err
		}

		album.Id = id
		return h.albumStorage.UpdateAlbum(&album)
	case http.MethodDelete:
		return nil, h.albumStorage.DeleteAlbum(id)
	default:
		return nil, http.ErrNotSupported
	}
}
