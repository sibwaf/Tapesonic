package handlers

import (
	"net/http"

	"tapesonic/api/admin/responses"
	"tapesonic/ytdlp"
)

type importHandler struct {
	ytdlp *ytdlp.Ytdlp
}

func NewImportHandler(
	ytdlp *ytdlp.Ytdlp,
) *importHandler {
	return &importHandler{
		ytdlp: ytdlp,
	}
}

func (h *importHandler) Methods() []string {
	return []string{http.MethodPost}
}

func (h *importHandler) Handle(r *http.Request) (*responses.Response, error) {
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

	metadata, err := h.ytdlp.Download(url, format, "data") // todo
	if err != nil {
		resp := responses.NewResponse(err) // todo
		return &resp, nil
	}

	resp := responses.NewResponse(metadata) // todo
	return &resp, nil
}
