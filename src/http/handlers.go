package http

import (
	"net/http"
	"path"

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

	handlers["/assets/"] = http.FileServer(http.Dir(appCtx.Config.WebappDir)).ServeHTTP
	handlers["/"] = func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(appCtx.Config.WebappDir, "index.html"))
	}

	return handlers
}
