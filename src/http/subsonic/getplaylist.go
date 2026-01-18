package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/subsonic"
)

func GetPlaylist(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		playlist, err := library.GetPlaylistById(user, id)
		if err != nil {
			return nil, err
		}

		playlistRs := ToPlaylist(playlist)

		response := subsonic.NewOkResponse()
		response.Playlist = &playlistRs
		return response, nil
	}
}
