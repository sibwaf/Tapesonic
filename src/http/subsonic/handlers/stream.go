package handlers

import (
	"io"
	"net/http"
	"time"

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

	stream, err := h.subsonic.Stream(r.Context(), id)
	if err != nil {
		return nil, err
	}

	defer stream.Reader.Close()

	if stream.MimeType != "" {
		w.Header().Add("Content-Type", stream.MimeType)
	}

	readSeeker, isSeekable := stream.Reader.(io.ReadSeeker)
	if isSeekable {
		http.ServeContent(w, r, "", time.Time{}, readSeeker)
	} else {
		io.Copy(w, stream.Reader)
	}

	return nil, nil
}
