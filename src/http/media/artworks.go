package media

import (
	"net/http"
	"tapesonic/library"
	"tapesonic/media"

	"github.com/gorilla/mux"
)

func GetArtwork(auth *authenticator, library *library.LibraryService, covers *media.ArtworkService) MediaHandler {
	return func(r *http.Request, w http.ResponseWriter) error {
		user, err := auth.Authenticate(r)
		if err != nil {
			return err
		}

		// TODO: ServeCover should fetch the artwork by id, not us
		artwork, err := library.GetArtwork(mux.Vars(r)["artworkId"])
		if err != nil {
			return err
		}

		return covers.ServeArtwork(user, r, w, artwork)
	}
}
