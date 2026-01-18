package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetPodcasts(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.Podcasts = &subsonic.Podcasts{
			Channel: []subsonic.PodcastChannel{},
		}
		return response, nil
	}
}
