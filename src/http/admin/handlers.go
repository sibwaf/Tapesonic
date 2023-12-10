package admin

import (
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/admin/handlers"
	"tapesonic/http/admin/util"

	"github.com/gorilla/mux"
)

func GetHandler(appCtx *appcontext.Context) (string, http.Handler) {
	// todo: logging
	rawHandlers := map[string]util.WebappHandler{
		"/api/formats": handlers.NewGetFormatsHandler(appCtx.Ytdlp),
		"/api/import":  handlers.NewImportHandler(appCtx.Importer),

		"/api/tapes":                  handlers.NewTapesHandler(appCtx.DataStorage),
		"/api/tapes/{tapeId}":         handlers.NewTapeHandler(appCtx.DataStorage),
		"/api/tapes/{tapeId}/related": handlers.NewTapeRelatedHandler(appCtx.DataStorage),

		"/api/playlists":                      handlers.NewPlaylistsHandler(appCtx.DataStorage),
		"/api/playlists/{playlistId}":         handlers.NewPlaylistHandler(appCtx.DataStorage),
		"/api/playlists/{playlistId}/related": handlers.NewPlaylistRelatedHandler(appCtx.DataStorage),
	}

	router := mux.NewRouter()
	for path, handler := range rawHandlers {
		router.HandleFunc(path, util.Authenticated(util.AsHandlerFunc(handler), appCtx.Config))
	}

	return "/api/", router
}
