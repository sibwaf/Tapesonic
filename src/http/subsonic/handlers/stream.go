package handlers

import (
	"errors"
	"io"
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type streamHandler struct {
	subsonic logic.SubsonicService
}

func NewStreamHandler(
	subsonic logic.SubsonicService,
) *streamHandler {
	return &streamHandler{
		subsonic: subsonic,
	}
}

func (h *streamHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	mimeType, reader, err := h.subsonic.Stream(r.Context(), id)
	if err != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", mimeType)
	_, err = io.Copy(w, reader)
	return nil, errors.Join(
		err,
		reader.Close(),
	)
}
