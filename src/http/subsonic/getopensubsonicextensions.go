package subsonic

import (
	"net/http"
	"tapesonic/subsonic"
)

func GetOpenSubsonicExtensions() SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		response := subsonic.NewOkResponse()
		response.OpenSubsonicExtensions = &subsonic.OpenSubsonicExtensions{
			OpenSubsonicExtensions: []subsonic.OpenSubsonicExtension{},
		}
		return response, nil
	}
}
