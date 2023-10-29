package util

import (
	"net/http"

	"tapesonic/config"
)

func Authenticated(handler http.HandlerFunc, config *config.TapesonicConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != config.Username || password != config.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
