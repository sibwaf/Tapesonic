package http

import (
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/admin"
	"tapesonic/http/subsonic"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc)

	for path, handler := range subsonic.GetHandlers(appCtx) {
		handlers[path] = handler
	}
	for path, handler := range admin.GetHandlers(appCtx) {
		handlers[path] = handler
	}

	return handlers
}
