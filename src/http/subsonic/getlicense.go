package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetLicense(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: check license on all connected remotes?

		response := subsonic.NewOkResponse()
		response.License = &subsonic.License{
			Valid: true,
		}
		return response, nil
	}
}
