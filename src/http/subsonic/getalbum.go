package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/subsonic"
)

func GetAlbum(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		id := r.URL.Query().Get("id")
		if id == "" {
			return subsonic.NewParameterMissingResponse("id"), nil
		}

		album, err := library.GetAlbumById(user, id)
		if err != nil {
			return nil, err
		}

		albumRs := ToAlbumId3(album)

		response := subsonic.NewOkResponse()
		response.Album = &albumRs
		return response, nil
	}
}
