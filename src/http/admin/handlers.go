package admin

import (
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/admin/handlers"
	"tapesonic/http/admin/util"

	"github.com/gorilla/mux"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	type PathHandler struct {
		Path    string
		Handler http.HandlerFunc
	}

	// todo: logging
	rawHandlers := []PathHandler{
		{Path: "/api/settings/lastfm/auth", Handler: util.AsHandlerFunc(handlers.NewSettingsLastFmAuthHandler(appCtx.LastFmService))},
		{Path: "/api/settings/lastfm/create-auth-link", Handler: util.AsHandlerFunc(handlers.NewSettingsLastFmCreateAuthLinkHandler(appCtx.LastFmService))},

		{Path: "/api/tapes", Handler: util.AsHandlerFunc(handlers.NewTapesHandler(appCtx.TapeService))},
		{Path: "/api/tapes/guess-metadata", Handler: util.AsHandlerFunc(handlers.NewGuessTapeMetadataHandler(appCtx.TapeService))},
		{Path: "/api/tapes/{tapeId}", Handler: util.AsHandlerFunc(handlers.NewTapeHandler(appCtx.TapeService))},

		{Path: "/api/sources", Handler: util.AsHandlerFunc(handlers.NewSourcesHandler(appCtx.SourceService))},
		{Path: "/api/sources/{sourceId}", Handler: util.AsHandlerFunc(handlers.NewSourceHandler(appCtx.SourceService))},
		{Path: "/api/sources/{sourceId}/hierarchy", Handler: util.AsHandlerFunc(handlers.NewSourceHierarchyHandler(appCtx.SourceService))},
		{Path: "/api/sources/{sourceId}/tracks", Handler: util.AsHandlerFunc(handlers.NewSourceTracksHandler(appCtx.TrackService))},
		{Path: "/api/sources/{sourceId}/file", Handler: util.AsHandlerFunc(handlers.NewSourceFileHandler(appCtx.SourceFileService))},

		{Path: "/api/tracks", Handler: util.AsHandlerFunc(handlers.NewTracksHandler(appCtx.TrackService, appCtx.SearchService))},

		{Path: "/api/thumbnails", Handler: util.AsHandlerFunc(handlers.NewThumbnailsHandler(appCtx.ThumbnailService))},

		{Path: "/media/thumbnails/{thumbnailId}", Handler: util.AsRawHandlerFunc(handlers.NewThumbnailRawHandler(appCtx.ThumbnailService))},
	}

	router := mux.NewRouter()
	for _, pathHandler := range rawHandlers {
		router.HandleFunc(pathHandler.Path, util.Authenticated(pathHandler.Handler, appCtx.Config))
	}

	// todo: wow that's disgusting
	return map[string]http.HandlerFunc{
		"/api/":   router.ServeHTTP,
		"/media/": router.ServeHTTP,
	}
}
