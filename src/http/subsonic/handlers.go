package subsonic

import (
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/subsonic/handlers"
	"tapesonic/http/subsonic/util"
)

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	rawHandlers := map[string]http.HandlerFunc{
		"/ping": util.AsHandlerFunc(handlers.Ping),

		"/getAlbumList2":            util.AsHandlerFunc(handlers.NewGetAlbumList2Handler(appCtx.SubsonicService).Handle),
		"/getAlbum":                 util.AsHandlerFunc(handlers.NewGetAlbumHandler(appCtx.SubsonicService).Handle),
		"/getArtists":               util.AsHandlerFunc(handlers.NewGetArtistsHandler().Handle),
		"/getGenres":                util.AsHandlerFunc(handlers.NewGetGenresHandler().Handle),
		"/getInternetRadioStations": util.AsHandlerFunc(handlers.NewGetInternetRadioStationsHandler().Handle),
		"/getMusicFolders":          util.AsHandlerFunc(handlers.NewGetMusicFoldersHandler().Handle),
		"/getNewestPodcasts":        util.AsHandlerFunc(handlers.NewGetNewestPodcastsHandler().Handle),
		"/getPlaylists":             util.AsHandlerFunc(handlers.NewGetPlaylistsHandler(appCtx.SubsonicService).Handle),
		"/getPlaylist":              util.AsHandlerFunc(handlers.NewGetPlaylistHandler(appCtx.SubsonicService).Handle),
		"/getPodcasts":              util.AsHandlerFunc(handlers.NewGetPodcastsHandler().Handle),
		"/getRandomSongs":           util.AsHandlerFunc(handlers.NewGetRandomSongsHandler().Handle),
		"/getScanStatus":            util.AsHandlerFunc(handlers.NewGetScanStatusHandler().Handle),
		"/getStarred2":              util.AsHandlerFunc(handlers.NewGetStarred2Handler().Handle),
		"/search3":                  util.AsHandlerFunc(handlers.NewSearch3Handler().Handle),

		"/scrobble": util.AsHandlerFunc(handlers.NewScrobbleHandler().Handle),

		"/stream":      util.AsRawHandlerFunc(handlers.NewStreamHandler(appCtx.SubsonicService).Handle),
		"/getCoverArt": util.AsRawHandlerFunc(handlers.NewGetCoverArtHandler(appCtx.SubsonicService).Handle),
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
