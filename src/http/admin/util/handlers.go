package util

import (
	"encoding/json"
	"net/http"
	"slices"

	"tapesonic/http/admin/responses"
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

		response, err := handler.Handle(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
