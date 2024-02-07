package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/storage"
)

type albumsHandler struct {
	albumStorage *storage.AlbumStorage
}

func NewAlbumsHandler(
	albumStorage *storage.AlbumStorage,
) *albumsHandler {
	return &albumsHandler{
		albumStorage: albumStorage,
	}
}

func (h *albumsHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *albumsHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		return h.albumStorage.GetAllAlbums()
	case http.MethodPost:
		var album storage.Album
		err := json.NewDecoder(r.Body).Decode(&album)
		if err != nil {
			return nil, err
		}

		return album, h.albumStorage.CreateAlbum(&album)
	default:
		return nil, http.ErrNotSupported
	}
}
