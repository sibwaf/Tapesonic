package media

import (
	"net/http"
	"tapesonic/library"
	"tapesonic/media"

	"github.com/gorilla/mux"
)

func GetThumbnail(auth *authenticator, library *library.LibraryService, covers *media.CoverService) MediaHandler {
	return func(r *http.Request, w http.ResponseWriter) error {
		user, err := auth.Authenticate(r)
		if err != nil {
			return err
		}

		// TODO: ServeCover should fetch the thumbnail by id, not us
		cover, err := library.GetCover(mux.Vars(r)["thumbnailId"])
		if err != nil {
			return err
		}

		return covers.ServeCover(user, r, w, cover)
	}
}
