package api

import (
	"net/http"

	"tapesonic/api/subsonic"
	"tapesonic/config"
)

func GetHandlers(config *config.TapesonicConfig) map[string]http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc)

	for path, handler := range subsonic.GetHandlers(config) {
		handlers[path] = handler
	}

	return handlers
}
