package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/media"
	"tapesonic/subsonic"
)

func GetCoverArt(auth *authenticator, library *library.LibraryService, covers *media.CoverService) SubsonicRawHandler {
	return func(w http.ResponseWriter, r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		cover, err := library.GetCover(id)
		if err != nil {
			return nil, err
		}

		return nil, covers.ServeCover(user, r, w, cover)
	}
}
