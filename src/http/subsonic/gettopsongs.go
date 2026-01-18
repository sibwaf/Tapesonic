package subsonic

import (
	"net/http"
	"tapesonic/subsonic"
)

func GetTopSongs(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.TopSongs = &subsonic.TopSongs{
			Song: []subsonic.Child{},
		}
		return response, nil
	}
}
