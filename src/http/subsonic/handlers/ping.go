package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

func Ping(r *http.Request) (*responses.SubsonicResponse, error) {
	return responses.NewOkResponse(), nil
}
