package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/media"
	"tapesonic/subsonic"
)

func Stream(auth *authenticator, library *library.LibraryService, streamer *media.StreamService) SubsonicRawHandler {
	return func(w http.ResponseWriter, r *http.Request) (*subsonic.Response, error) {
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

		return nil, streamer.ServeStream(user, r, w, track)
	}
}
