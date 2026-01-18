package http

import (
	"fmt"
	"log/slog"
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

	for path, handler := range admin.GetHandlers(appCtx) {
		handlers[path] = handler
	}

	handlers["/assets/"] = http.FileServer(http.Dir(appCtx.Config.WebappDir)).ServeHTTP
	handlers["/"] = func(w http.ResponseWriter, r *http.Request) {
		adminExists, err := appCtx.Users.UserService.CheckAdminAccountExists()
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to check if admin account exists: %s", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !adminExists && r.URL.Path != "/setup" {
			w.Header().Add("Location", "/setup")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		http.ServeFile(w, r, path.Join(appCtx.Config.WebappDir, "index.html"))
	}

	if appCtx.Config.DevMode {
		handlers["/debug/pprof/"] = pprof.Index
		handlers["/debug/pprof/cmdline"] = pprof.Cmdline
		handlers["/debug/pprof/profile"] = pprof.Profile
		handlers["/debug/pprof/symbol"] = pprof.Symbol
		handlers["/debug/pprof/trace"] = pprof.Trace
	}

	return handlers
}
