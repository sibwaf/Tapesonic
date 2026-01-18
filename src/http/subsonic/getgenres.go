package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetGenres(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.Genres = &subsonic.Genres{
			Genre: []subsonic.Genre{},
		}
		return response, nil
	}
}
