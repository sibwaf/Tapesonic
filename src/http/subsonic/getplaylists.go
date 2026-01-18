package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/subsonic"
	"tapesonic/util"
)

func GetPlaylists(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		playlists, err := library.GetPlaylists(user)
		if err != nil {
			return nil, err
		}

		response := subsonic.NewOkResponse()
		response.Playlists = &subsonic.Playlists{
			Playlist: util.Map(playlists, ToPlaylist),
		}
		return response, nil
	}
}
