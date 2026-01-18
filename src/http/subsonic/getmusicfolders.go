package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetMusicFolders(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.MusicFolders = &subsonic.MusicFolders{
			MusicFolder: []subsonic.MusicFolder{},
		}
		return response, nil
	}
}
