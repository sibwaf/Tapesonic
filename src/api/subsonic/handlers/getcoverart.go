package handlers

import (
	"io"
	"net/http"
	"os"

	"tapesonic/api/subsonic/responses"
	"tapesonic/storage"
)

type getCoverArtHandler struct {
	storage *storage.Storage
}

func NewGetCoverArtHandler(
	storage *storage.Storage,
) *getCoverArtHandler {
	return &getCoverArtHandler{
		storage: storage,
	}
}

func (h *getCoverArtHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	cover, err := h.storage.GetCover(id)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(cover.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", "image/png")
	_, err = io.Copy(w, reader)
	return nil, err
}
