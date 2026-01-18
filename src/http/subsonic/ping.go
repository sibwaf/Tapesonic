package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func Ping(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		return subsonic.NewOkResponse(), nil
	}
}
