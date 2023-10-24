package admin

import (
	"net/http"

	"tapesonic/api/admin/handlers"
	"tapesonic/api/admin/util"
	"tapesonic/appcontext"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	// todo: logging
	// todo: auth
	handlers := map[string]http.HandlerFunc{
		"/api/formats": util.AsHandlerFunc(handlers.NewGetFormatsHandler(appCtx.Ytdlp)),
		"/api/import":  util.AsHandlerFunc(handlers.NewImportHandler(appCtx.Ytdlp)),
	}

	return handlers
}
