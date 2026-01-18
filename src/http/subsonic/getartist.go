package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/subsonic"
)

func GetArtist(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		artist, err := library.GetArtist(user, id)
		if err != nil {
			return nil, err
		}

		artistRs := ToArtistId3(artist)

		response := subsonic.NewOkResponse()
		response.Artist = &artistRs
		return response, nil
	}
}
