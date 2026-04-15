package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/media"
	"tapesonic/subsonic"
)

func GetCoverArt(auth *authenticator, library *library.LibraryService, artworks *media.ArtworkService) SubsonicRawHandler {
	return func(w http.ResponseWriter, r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		artwork, err := library.GetArtwork(id)
		if err != nil {
			return nil, err
		}

		return nil, artworks.ServeArtwork(user, r, w, artwork)
	}
}
