package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetNewestPodcasts(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.NewestPodcasts = &subsonic.NewestPodcasts{
			Episode: []subsonic.PodcastEpisode{},
		}
		return response, nil
	}
}
