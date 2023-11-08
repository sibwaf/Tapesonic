package util

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
)

type WebappHandler interface {
	Methods() []string
	Handle(r *http.Request) (response any, err error)
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
			// todo
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error(err.Error())
			return
		}

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
