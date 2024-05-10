package handlers

import (
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
	"tapesonic/util"
)

type scrobbleHandler struct {
	subsonic logic.SubsonicService
}

func NewScrobbleHandler(subsonic logic.SubsonicService) *scrobbleHandler {
	return &scrobbleHandler{subsonic: subsonic}
}

func (h *scrobbleHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	timeStr := r.URL.Query().Get("time")
	time_ := time.UnixMilli(util.StringToInt64OrDefault(timeStr, time.Now().UnixMilli()))

	submissionStr := r.URL.Query().Get("submission")
	submission := util.StringToBoolOrDefault(submissionStr, true)

	return responses.NewOkResponse(), h.subsonic.Scrobble(id, time_, submission)
}
