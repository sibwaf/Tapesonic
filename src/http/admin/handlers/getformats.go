package handlers

import (
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/ytdlp"
)

type getFormatsHandler struct {
	ytdlp *ytdlp.Ytdlp
}

func NewGetFormatsHandler(
	ytdlp *ytdlp.Ytdlp,
) *getFormatsHandler {
	return &getFormatsHandler{
		ytdlp: ytdlp,
	}
}

func (h *getFormatsHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *getFormatsHandler) Handle(r *http.Request) (any, error) {
	url := r.URL.Query().Get("url")
	if url == "" {
		resp := responses.NewResponse("no url") // todo
		return &resp, nil
	}

	metadata, err := h.ytdlp.ExtractMetadata(url)
	if err != nil {
		resp := responses.NewResponse(err) // todo
		return &resp, nil
	}

	resp := responses.NewResponse(metadata) // todo
	return &resp, nil
}
