package handlers

import (
	"net/http"
	"os"

	"tapesonic/ffmpeg"
	"tapesonic/http/subsonic/responses"
	"tapesonic/http/util"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type streamHandler struct {
	mediaStorage *storage.MediaStorage
	ffmpeg       *ffmpeg.Ffmpeg
}

func NewStreamHandler(
	mediaStorage *storage.MediaStorage,
	ffmpeg *ffmpeg.Ffmpeg,
) *streamHandler {
	return &streamHandler{
		mediaStorage: mediaStorage,
		ffmpeg:       ffmpeg,
	}
}

func (h *streamHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	rawId := r.URL.Query().Get("id")
	if rawId == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	track, err := h.mediaStorage.GetTrack(id)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(track.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", util.FormatToMediaType(track.Format))
	return nil, h.ffmpeg.Stream(
		r.Context(),
		track.StartOffsetMs,
		track.EndOffsetMs-track.StartOffsetMs,
		ffmpeg.NewReaderWithMeta(
			"file://"+track.Path,
			reader,
		),
		w,
	)
}
