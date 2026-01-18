package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetArtists(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.Artists = &subsonic.Artists{
			IgnoredArticles: "",
			Index:           []subsonic.IndexId3{},
		}
		return response, nil
	}
}
