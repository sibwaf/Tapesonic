package subsonic

import (
	"net/http"

	"tapesonic/api/subsonic/handlers"
	"tapesonic/api/subsonic/util"
	"tapesonic/config"
)

func GetHandlers(config *config.TapesonicConfig) map[string]http.HandlerFunc {
	rawHandlers := map[string]http.HandlerFunc{
		"/rest/ping": util.AsHandlerFunc(handlers.Ping),
	}

	handlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		wrappedHandler := util.Logged(util.Authenticated(handler, config))
		handlers[path] = wrappedHandler
		handlers[path+".view"] = wrappedHandler
	}

	return handlers
}
