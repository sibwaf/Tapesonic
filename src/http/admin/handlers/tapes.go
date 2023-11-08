package handlers

import (
	"net/http"

	"tapesonic/storage"
)

type tapesHandler struct {
	dataStorage *storage.DataStorage
}

func NewTapesHandler(
	dataStorage *storage.DataStorage,
) *tapesHandler {
	return &tapesHandler{
		dataStorage: dataStorage,
	}
}

func (h *tapesHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *tapesHandler) Handle(r *http.Request) (any, error) {
	return h.dataStorage.GetAllTapes()
}
