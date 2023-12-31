package handlers

import (
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/storage"
)

type importHandler struct {
	importer *storage.Importer
}

func NewImportHandler(
	importer *storage.Importer,
) *importHandler {
	return &importHandler{
		importer: importer,
	}
}

func (h *importHandler) Methods() []string {
	return []string{http.MethodPost}
}

func (h *importHandler) Handle(r *http.Request) (any, error) {
	url := r.URL.Query().Get("url")
	if url == "" {
		resp := responses.NewResponse("no url") // todo
		return &resp, nil
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		resp := responses.NewResponse("no format") // todo
		return &resp, nil
	}

	id, err := h.importer.ImportTape(url, format)
	if err != nil {
		return nil, err
	}

	resp := responses.NewResponse(id)
	return &resp, nil // todo
}
