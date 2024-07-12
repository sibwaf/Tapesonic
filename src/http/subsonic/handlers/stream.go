package handlers

import (
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type streamHandler struct {
	streamService *logic.StreamService
}

func NewStreamHandler(
	streamService *logic.StreamService,
) *streamHandler {
	return &streamHandler{
		streamService: streamService,
	}
}

func (h *streamHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	mediaType, reader, err := h.streamService.Stream(r.Context(), id)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	if mediaType != "" {
		w.Header().Add("Content-Type", mediaType)
	}

	http.ServeContent(w, r, id, time.Time{}, reader)

	return nil, nil
}
