package subsonic

import (
	"net/http"
	"time"

	"tapesonic/scrobbling"
	"tapesonic/subsonic"
	"tapesonic/util"
)

func Scrobble(auth *authenticator, scrobbler *scrobbling.ScrobbleService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		timeStr := r.URL.Query().Get("time")
		time_ := time.UnixMilli(util.StringToInt64OrDefault(timeStr, time.Now().UnixMilli()))

		submissionStr := r.URL.Query().Get("submission")
		submission := util.StringToBoolOrDefault(submissionStr, true)

		if submission {
			return subsonic.NewOkResponse(), scrobbler.ScrobbleCompleted(user, id, time_)
		} else {
			return subsonic.NewOkResponse(), scrobbler.ScrobblePlaying(user, id, time_)
		}
	}
}
