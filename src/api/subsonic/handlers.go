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

		"/getAlbumList2":            util.AsHandlerFunc(handlers.NewGetAlbumList2Handler().Handle),
		"/getArtists":               util.AsHandlerFunc(handlers.NewGetArtistsHandler().Handle),
		"/getGenres":                util.AsHandlerFunc(handlers.NewGetGenresHandler().Handle),
		"/getInternetRadioStations": util.AsHandlerFunc(handlers.NewGetInternetRadioStationsHandler().Handle),
		"/getMusicFolders":          util.AsHandlerFunc(handlers.NewGetMusicFoldersHandler().Handle),
		"/getNewestPodcasts":        util.AsHandlerFunc(handlers.NewGetNewestPodcastsHandler().Handle),
		"/getPlaylists":             util.AsHandlerFunc(handlers.NewGetPlaylistsHandler(appCtx.Storage).Handle),
		"/getPlaylist":              util.AsHandlerFunc(handlers.NewGetPlaylistHandler(appCtx.Storage).Handle),
		"/getPodcasts":              util.AsHandlerFunc(handlers.NewGetPodcastsHandler().Handle),
		"/getRandomSongs":           util.AsHandlerFunc(handlers.NewGetRandomSongsHandler().Handle),
		"/getScanStatus":            util.AsHandlerFunc(handlers.NewGetScanStatusHandler().Handle),
		"/getStarred2":              util.AsHandlerFunc(handlers.NewGetStarred2Handler().Handle),
		"/search3":                  util.AsHandlerFunc(handlers.NewSearch3Handler().Handle),

		"/scrobble": util.AsHandlerFunc(handlers.NewScrobbleHandler().Handle),

		"/stream":      util.AsRawHandlerFunc(handlers.NewStreamHandler(appCtx.Storage, appCtx.Ffmpeg).Handle),
		"/getCoverArt": util.AsRawHandlerFunc(handlers.NewGetCoverArtHandler(appCtx.Storage).Handle),
	}

	resultHandlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		wrappedHandler := util.Logged(util.Authenticated(handler, appCtx.Config))
		resultHandlers["/rest"+path] = wrappedHandler
		resultHandlers["/rest"+path+".view"] = wrappedHandler
	}

	resultHandlers["/rest/"] = util.Logged(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			util.LogWarning(r, "Handler is not implemented")
		},
	)

	return resultHandlers
}
