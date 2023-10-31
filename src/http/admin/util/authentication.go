package util

import (
	"net/http"

	"tapesonic/config"
)

func Authenticated(handler http.HandlerFunc, config *config.TapesonicConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != config.Username || password != config.Password {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"master\", charset=\"UTF-8\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
