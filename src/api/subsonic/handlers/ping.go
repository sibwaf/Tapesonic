package handlers

import (
	"net/http"

	"tapesonic/api/subsonic/responses"
)

func Ping(r *http.Request) (*responses.SubsonicResponse, error) {
	return responses.NewOkSubsonicResponse(), nil
}
