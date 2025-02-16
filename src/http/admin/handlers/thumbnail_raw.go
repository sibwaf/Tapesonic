package handlers

import (
	"fmt"
	"io"
	"net/http"

	"tapesonic/logic"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type thumbnailRawHandler struct {
	service *logic.ThumbnailService
}

func NewThumbnailRawHandler(
	service *logic.ThumbnailService,
) *thumbnailRawHandler {
	return &thumbnailRawHandler{
		service: service,
	}
}

func (h *thumbnailRawHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *thumbnailRawHandler) Handle(r *http.Request, w http.ResponseWriter) error {
	tapeId, idErr := uuid.Parse(mux.Vars(r)["thumbnailId"])
	if idErr != nil {
		return fmt.Errorf("missing or invalid thumbnailId")
	}

	// todo: ServeFile
	mediaType, reader, err := h.service.GetThumbnailContent(tapeId)
	if err != nil {
		return err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", mediaType)
	_, err = io.Copy(w, reader)

	return err
}
