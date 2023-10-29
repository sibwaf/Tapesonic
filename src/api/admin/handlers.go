package admin

import (
	"net/http"

	"tapesonic/api/admin/handlers"
	"tapesonic/api/admin/util"
	"tapesonic/appcontext"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	// todo: logging
	rawHandlers := map[string]util.WebappHandler{
		"/api/formats": handlers.NewGetFormatsHandler(appCtx.Ytdlp),
		"/api/import":  handlers.NewImportHandler(appCtx.Ytdlp),
	}

	resultHandlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		resultHandlers[path] = util.Authenticated(util.AsHandlerFunc(handler), appCtx.Config)
	}

	return resultHandlers
}
