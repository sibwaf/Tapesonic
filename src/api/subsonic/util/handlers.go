package util

import (
	"net/http"

	"tapesonic/api/subsonic/responses"
)

type SubsonicHandler func(r *http.Request) (response *responses.SubsonicResponse, err error)

func AsHandlerFunc(handler SubsonicHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(r)
		if err != nil {
			LogError(r, "Failed to process request", "error", err)
			response = responses.NewFailureSubsonicResponse(responses.ERROR_CODE_GENERIC, "Server failed to process the request")
		}

		writeResponse(w, r, response)
	}
}
