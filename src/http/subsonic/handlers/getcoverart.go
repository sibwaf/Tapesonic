package handlers

import (
	"io"
	"net/http"
	"os"

	"tapesonic/http/subsonic/responses"
	"tapesonic/http/util"
	"tapesonic/storage"
)

type getCoverArtHandler struct {
	mediaStorage *storage.MediaStorage
}

func NewGetCoverArtHandler(
	mediaStorage *storage.MediaStorage,
) *getCoverArtHandler {
	return &getCoverArtHandler{
		mediaStorage: mediaStorage,
	}
}

func (h *getCoverArtHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	cover, err := h.mediaStorage.GetCover(id)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(cover.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", util.FormatToMediaType(cover.Format))
	_, err = io.Copy(w, reader)
	return nil, err
}
