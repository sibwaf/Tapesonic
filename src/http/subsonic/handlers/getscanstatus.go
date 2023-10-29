package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getScanStatusHandler struct {
}

func NewGetScanStatusHandler() *getScanStatusHandler {
	return &getScanStatusHandler{}
}

func (h *getScanStatusHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.ScanStatus = responses.NewScanStatus(false, 0)
	return response, nil
}
