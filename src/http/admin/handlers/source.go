package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type sourceHandler struct {
	sources *logic.SourceService
}

func NewSourceHandler(
	sources *logic.SourceService,
) *sourceHandler {
	return &sourceHandler{
		sources: sources,
	}
}

func (h *sourceHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *sourceHandler) Handle(r *http.Request) (any, error) {
	sourceId, idErr := uuid.Parse(mux.Vars(r)["sourceId"])
	if idErr != nil {
		return nil, fmt.Errorf("missing or invalid sourceId")
	}

	switch r.Method {
	case http.MethodGet:
		source, err := h.sources.GetById(sourceId)
		if err != nil {
			return nil, err
		}

		return responses.SourceToFullSourceRs(source), nil
	default:
		return nil, http.ErrNotSupported
	}
}
