package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tapesonic/http/admin/requests"
	"tapesonic/http/admin/responses"
	"tapesonic/logic"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type tapeHandler struct {
	service *logic.TapeService
}

func NewTapeHandler(
	service *logic.TapeService,
) *tapeHandler {
	return &tapeHandler{
		service: service,
	}
}

func (h *tapeHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPut, http.MethodDelete}
}

func (h *tapeHandler) Handle(r *http.Request) (any, error) {
	tapeId, idErr := uuid.Parse(mux.Vars(r)["tapeId"])
	if idErr != nil {
		return nil, fmt.Errorf("missing or invalid tapeId")
	}

	switch r.Method {
	case http.MethodGet:
		tape, tracks, err := h.service.GetById(tapeId)
		if err != nil {
			return nil, err
		}

		return responses.TapeToDto(tape, tracks), nil
	case http.MethodPut:
		var tapeRequest requests.ModifiedTape
		err := json.NewDecoder(r.Body).Decode(&tapeRequest)
		if err != nil {
			return nil, err
		}

		if tapeRequest.Id != tapeId {
			return nil, fmt.Errorf("tapeId mismatch: tapeId=%s, tape.id=%s", tapeId, tapeRequest.Id)
		}

		tape := requests.ModifiedTapeToModel(tapeRequest)

		tape, tracks, err := h.service.Update(tape)
		if err != nil {
			return nil, err
		}

		return responses.TapeToDto(tape, tracks), nil
	case http.MethodDelete:
		return nil, h.service.DeleteById(tapeId)
	default:
		return nil, http.ErrNotSupported
	}
}
