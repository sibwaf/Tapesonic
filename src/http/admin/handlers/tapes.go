package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/http/admin/requests"
	"tapesonic/http/admin/responses"
	"tapesonic/logic"
)

type tapesHandler struct {
	service *logic.TapeService
}

func NewTapesHandler(
	service *logic.TapeService,
) *tapesHandler {
	return &tapesHandler{
		service: service,
	}
}

func (h *tapesHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *tapesHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		tapes, err := h.service.GetList()
		if err != nil {
			return nil, err
		}

		return responses.TapesToListDto(tapes), nil
	case http.MethodPost:
		var tapeRequest requests.ModifiedTape
		err := json.NewDecoder(r.Body).Decode(&tapeRequest)
		if err != nil {
			return nil, err
		}

		tape := requests.ModifiedTapeToModel(tapeRequest)

		tape, tracks, err := h.service.Create(tape)
		if err != nil {
			return nil, err
		}

		return responses.TapeToDto(tape, tracks), nil
	default:
		return nil, http.ErrNotSupported
	}
}
