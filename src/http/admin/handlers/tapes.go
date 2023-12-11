package handlers

import (
	"net/http"

	"tapesonic/storage"
)

type tapesHandler struct {
	tapeStorage *storage.TapeStorage
}

func NewTapesHandler(
	tapeStorage *storage.TapeStorage,
) *tapesHandler {
	return &tapesHandler{
		tapeStorage: tapeStorage,
	}
}

func (h *tapesHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *tapesHandler) Handle(r *http.Request) (any, error) {
	return h.tapeStorage.GetAllTapes()
}
