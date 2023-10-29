package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getInternetRadioStationsHandler struct {
}

func NewGetInternetRadioStationsHandler() *getInternetRadioStationsHandler {
	return &getInternetRadioStationsHandler{}
}

func (h *getInternetRadioStationsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.InternetRadioStations = responses.NewInternetRadioStations([]responses.InternetRadioStation{})
	return response, nil
}
