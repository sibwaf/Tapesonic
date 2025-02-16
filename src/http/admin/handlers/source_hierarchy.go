package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type sourceHierarchyHandler struct {
	sources *logic.SourceService
}

func NewSourceHierarchyHandler(
	sources *logic.SourceService,
) *sourceHierarchyHandler {
	return &sourceHierarchyHandler{
		sources: sources,
	}
}

func (h *sourceHierarchyHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *sourceHierarchyHandler) Handle(r *http.Request) (any, error) {
	sourceId, idErr := uuid.Parse(mux.Vars(r)["sourceId"])
	if idErr != nil {
		return nil, fmt.Errorf("missing or invalid sourceId")
	}

	switch r.Method {
	case http.MethodGet:
		hierarchy, err := h.sources.GetHierarchy(sourceId)
		if err != nil {
			return nil, err
		}

		return responses.SourcesForHierarchyToListSourceHierarchyRs(hierarchy), nil
	default:
		return nil, http.ErrNotSupported
	}
}
