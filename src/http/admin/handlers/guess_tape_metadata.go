package handlers

import (
	"encoding/json"
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type GuessTapeMetadataRq struct {
	TrackIds []uuid.UUID
}

type guessTapeMetadataHandler struct {
	tapes *logic.TapeService
}

func NewGuessTapeMetadataHandler(
	tapes *logic.TapeService,
) *guessTapeMetadataHandler {
	return &guessTapeMetadataHandler{
		tapes: tapes,
	}
}

func (h *guessTapeMetadataHandler) Methods() []string {
	return []string{http.MethodPost}
}

func (h *guessTapeMetadataHandler) Handle(r *http.Request) (any, error) {
	request := GuessTapeMetadataRq{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	guessedTape, err := h.tapes.GuessTapeMetadata(request.TrackIds)
	if err != nil {
		return nil, err
	}

	return responses.TapeToDto(guessedTape, []storage.Track{}), nil
}
