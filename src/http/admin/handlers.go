package admin

import (
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/admin/handlers"
	"tapesonic/http/admin/util"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	// todo: logging
	rawHandlers := map[string]util.WebappHandler{
		"/api/formats": handlers.NewGetFormatsHandler(appCtx.Ytdlp),
		"/api/import":  handlers.NewImportHandler(appCtx.Importer),
	}

	resultHandlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		resultHandlers[path] = util.Authenticated(util.AsHandlerFunc(handler), appCtx.Config)
	}

	return resultHandlers
}
