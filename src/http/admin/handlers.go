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

		"/api/import-queue":          handlers.NewImportQueueHandler(appCtx.ImportQueueStorage),
		"/api/import-queue/{itemId}": handlers.NewImportQueueItemHandler(appCtx.ImportQueueStorage),

		"/api/tapes":                  handlers.NewTapesHandler(appCtx.TapeStorage),
		"/api/tapes/{tapeId}":         handlers.NewTapeHandler(appCtx.TapeStorage),
		"/api/tapes/{tapeId}/related": handlers.NewTapeRelatedHandler(appCtx.TapeStorage),

		"/api/playlists":                      handlers.NewPlaylistsHandler(appCtx.PlaylistStorage),
		"/api/playlists/{playlistId}":         handlers.NewPlaylistHandler(appCtx.PlaylistStorage),
		"/api/playlists/{playlistId}/related": handlers.NewPlaylistRelatedHandler(appCtx.PlaylistStorage),

		"/api/albums":                   handlers.NewAlbumsHandler(appCtx.AlbumStorage),
		"/api/albums/{albumId}":         handlers.NewAlbumHandler(appCtx.AlbumStorage),
		"/api/albums/{albumId}/related": handlers.NewAlbumRelatedHandler(appCtx.AlbumStorage),
	}

	router := mux.NewRouter()
	for path, handler := range rawHandlers {
		router.HandleFunc(path, util.Authenticated(util.AsHandlerFunc(handler), appCtx.Config))
	}

	return "/api/", router
}
