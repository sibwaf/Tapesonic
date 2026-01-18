package subsonic

import (
	"net/http"
	"tapesonic/library"
	"tapesonic/subsonic"
)

func GetSong(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		track, err := library.GetTrackById(user.Id, id)
		if err != nil {
			return nil, err
		}

		trackRs := ToChild(track)

		response := subsonic.NewOkResponse()
		response.Song = &trackRs
		return response, nil
	}
}
