package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type sourceDownloadHandler struct {
	files *logic.SourceFileService
}

func NewSourceFileHandler(
	files *logic.SourceFileService,
) *sourceDownloadHandler {
	return &sourceDownloadHandler{
		files: files,
	}
}

func (h *sourceDownloadHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodDelete}
}

func (h *sourceDownloadHandler) Handle(r *http.Request) (any, error) {
	sourceId, idErr := uuid.Parse(mux.Vars(r)["sourceId"])
	if idErr != nil {
		return nil, fmt.Errorf("missing or invalid sourceId")
	}

	switch r.Method {
	case http.MethodGet:
		file, err := h.files.FindBySourceId(sourceId)
		if err != nil {
			return nil, err
		}

		if file == nil {
			// todo: 404
			return nil, nil
		} else {
			return responses.SourceFileToSourceFileRs(*file), nil
		}
	case http.MethodDelete:
		err := h.files.DeleteFor(sourceId)
		if err != nil {
			return nil, err
		}

		return nil, nil
	default:
		return nil, http.ErrNotSupported
	}
}
