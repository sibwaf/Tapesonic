package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetStarred2(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.Starred2 = &subsonic.Starred2{
			Artist: []subsonic.ArtistId3{},
			Album:  []subsonic.AlbumId3{},
			Song:   []subsonic.Child{},
		}
		return response, nil
	}
}
