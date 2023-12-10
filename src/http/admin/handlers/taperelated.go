package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type tapeRelatedHandler struct {
	dataStorage *storage.DataStorage
}

func NewTapeRelatedHandler(
	dataStorage *storage.DataStorage,
) *tapeRelatedHandler {
	return &tapeRelatedHandler{
		dataStorage: dataStorage,
	}
}

func (h *tapeRelatedHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *tapeRelatedHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["tapeId"]
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	return h.dataStorage.GetTapeRelationships(id)
}
