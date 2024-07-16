package http

import (
	"net/http"
	"path"

	"tapesonic/appcontext"
	"tapesonic/http/admin"
	"tapesonic/http/subsonic"

	"net/http/pprof"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc)

	for path, handler := range subsonic.GetHandlers(appCtx) {
		handlers[path] = handler
	}

	apiPath, apiHandler := admin.GetHandler(appCtx)
	handlers[apiPath] = apiHandler.ServeHTTP

	handlers["/assets/"] = http.FileServer(http.Dir(appCtx.Config.WebappDir)).ServeHTTP
	handlers["/"] = func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(appCtx.Config.WebappDir, "index.html"))
	}

	if appCtx.Config.DevMode {
		handlers["/debug/pprof"] = pprof.Index
		handlers["/debug/pprof/cmdline"] = pprof.Cmdline
		handlers["/debug/pprof/profile"] = pprof.Profile
		handlers["/debug/pprof/symbol"] = pprof.Symbol
		handlers["/debug/pprof/trace"] = pprof.Trace
	}

	return handlers
}
