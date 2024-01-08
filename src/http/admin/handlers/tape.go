package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type tapeHandler struct {
	tapeStorage *storage.TapeStorage
}

func NewTapeHandler(
	tapeStorage *storage.TapeStorage,
) *tapeHandler {
	return &tapeHandler{
		tapeStorage: tapeStorage,
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
		return h.tapeStorage.GetTapeWithFilesAndTracks(id)
	case http.MethodPut:
		var tape storage.Tape
		err := json.NewDecoder(r.Body).Decode(&tape)
		if err != nil {
			return nil, err
		}

		return nil, h.tapeStorage.UpsertTape(&tape)
	default:
		return nil, http.ErrNotSupported
	}
}
