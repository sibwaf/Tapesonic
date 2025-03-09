package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
)

type getLicenseHandler struct {
	subsonic logic.SubsonicService
}

func NewGetLicenseHandler(
	subsonic logic.SubsonicService,
) *getLicenseHandler {
	return &getLicenseHandler{
		subsonic: subsonic,
	}
}

func (h *getLicenseHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	license, err := h.subsonic.GetLicense()
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.License = license
	return response, nil
}
