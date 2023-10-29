package handlers

import (
	"net/http"
	"os"

	"tapesonic/api/subsonic/responses"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
)

type streamHandler struct {
	storage *storage.Storage
	ffmpeg  *ffmpeg.Ffmpeg
}

func NewStreamHandler(
	storage *storage.Storage,
	ffmpeg *ffmpeg.Ffmpeg,
) *streamHandler {
	return &streamHandler{
		storage: storage,
		ffmpeg:  ffmpeg,
	}
}

func (h *streamHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	track, err := h.storage.GetStreamableTrack(id)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(track.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", "audio/opus")
	return nil, h.ffmpeg.Stream(
		r.Context(),
		track.Track.OffsetMs,
		track.Track.LengthMs,
		ffmpeg.NewReaderWithMeta(
			"file:"+track.Path,
			reader,
		),
		w,
	)
}
