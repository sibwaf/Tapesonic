package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type tapeHandler struct {
	dataStorage *storage.DataStorage
}

func NewTapeHandler(
	dataStorage *storage.DataStorage,
) *tapeHandler {
	return &tapeHandler{
		dataStorage: dataStorage,
	}
}

func (h *tapeHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPut}
}

func (h *tapeHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["tapeId"]
	id, idErr := uuid.Parse(rawId)

	switch r.Method {
	case http.MethodGet:
		if idErr != nil {
			return nil, idErr
		}
		return h.dataStorage.GetTapeWithTracks(id)
	case http.MethodPut:
		var tape storage.Tape
		err := json.NewDecoder(r.Body).Decode(&tape)
		if err != nil {
			return nil, err
		}

		return nil, h.dataStorage.UpsertTape(&tape)
	default:
		return nil, http.ErrNotSupported
	}
}
