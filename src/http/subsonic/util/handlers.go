package util

import (
	"fmt"
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type SubsonicHandler func(r *http.Request) (response *responses.SubsonicResponse, err error)

type SubsonicRawHandler func(w http.ResponseWriter, r *http.Request) (response *responses.SubsonicResponse, err error)

func AsHandlerFunc(handler SubsonicHandler) http.HandlerFunc {
	return AsRawHandlerFunc(func(w http.ResponseWriter, r *http.Request) (response *responses.SubsonicResponse, err error) {
		return handler(r)
	})
}

func AsRawHandlerFunc(handler SubsonicRawHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(w, r)
		if err != nil {
			LogError(r, fmt.Sprintf("Failed to process request: %s", err.Error()))
			response = responses.NewFailedResponse(responses.ERROR_CODE_GENERIC, "Server failed to process the request")
		}

		if response != nil {
			writeResponse(w, r, response)
		}
	}
}
