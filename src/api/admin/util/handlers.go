package util

import (
	"encoding/json"
	"net/http"

	"tapesonic/api/admin/responses"
)

type WebappHandler interface {
	Handle(r *http.Request) (response *responses.Response, err error)
}

func AsHandlerFunc(handler WebappHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, _ := handler.Handle(r)
		// todo: err

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
