package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/subsonic"
	"tapesonic/util"
)

func GetRandomSongs(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		size := util.StringToIntOrDefault(r.URL.Query().Get("size"), 10)
		size = max(size, 0)
		size = min(size, 500)

		// genre := r.URL.Query().Get("genre") // todo

		fromYear := util.StringToIntOrNull(r.URL.Query().Get("fromYear"))
		toYear := util.StringToIntOrNull(r.URL.Query().Get("toYear"))

		tracks, err := library.GetRandomTracks(user, size, fromYear, toYear)
		if err != nil {
			return nil, err
		}

		response := subsonic.NewOkResponse()
		response.RandomSongs = &subsonic.RandomSongs{
			Song: util.Map(tracks, ToChild),
		}
		return response, nil
	}
}
