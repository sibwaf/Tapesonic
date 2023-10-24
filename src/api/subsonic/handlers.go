package subsonic

import (
	"net/http"

	"tapesonic/api/subsonic/handlers"
	"tapesonic/api/subsonic/util"
	"tapesonic/appcontext"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	rawHandlers := map[string]http.HandlerFunc{
		"/ping": util.AsHandlerFunc(handlers.Ping),

		"/getPlaylists": util.AsHandlerFunc(handlers.NewGetPlaylistsHandler(appCtx.Storage).Handle),
		"/getPlaylist":  util.AsHandlerFunc(handlers.NewGetPlaylistHandler(appCtx.Storage).Handle),

		"/stream":      util.AsRawHandlerFunc(handlers.NewStreamHandler(appCtx.Storage, appCtx.Ffmpeg).Handle),
		"/getCoverArt": util.AsRawHandlerFunc(handlers.NewGetCoverArtHandler(appCtx.Storage).Handle),
	}

	resultHandlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		wrappedHandler := util.Logged(util.Authenticated(handler, appCtx.Config))
		resultHandlers["/rest"+path] = wrappedHandler
		resultHandlers["/rest"+path+".view"] = wrappedHandler
	}

	return resultHandlers
}
