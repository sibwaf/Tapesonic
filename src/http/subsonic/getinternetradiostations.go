package subsonic

import (
	"net/http"

	"tapesonic/subsonic"
)

func GetInternetRadioStations(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.InternetRadioStations = &subsonic.InternetRadioStations{
			InternetRadioStation: []subsonic.InternetRadioStation{},
		}
		return response, nil
	}
}
