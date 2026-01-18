package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetScanStatus(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.ScanStatus = &subsonic.ScanStatus{
			Scanning: false,
			Count:    0,
		}
		return response, nil
	}
}
