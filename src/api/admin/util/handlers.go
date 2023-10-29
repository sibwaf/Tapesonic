package util

import (
	"encoding/json"
	"net/http"
	"slices"

	"tapesonic/api/admin/responses"
)

type WebappHandler interface {
	Methods() []string
	Handle(r *http.Request) (response *responses.Response, err error)
}

func AsHandlerFunc(handler WebappHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedMethods := handler.Methods()
		if len(allowedMethods) > 0 && !slices.Contains(allowedMethods, r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		response, _ := handler.Handle(r)
		// todo: err

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
