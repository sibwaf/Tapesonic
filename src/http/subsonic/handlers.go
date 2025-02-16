package subsonic

import (
	"fmt"
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
		"/getArtist":                util.AsHandlerFunc(handlers.NewGetArtistHandler(appCtx.SubsonicService).Handle),
		"/getGenres":                util.AsHandlerFunc(handlers.NewGetGenresHandler().Handle),
		"/getInternetRadioStations": util.AsHandlerFunc(handlers.NewGetInternetRadioStationsHandler().Handle),
		"/getMusicFolders":          util.AsHandlerFunc(handlers.NewGetMusicFoldersHandler().Handle),
		"/getNewestPodcasts":        util.AsHandlerFunc(handlers.NewGetNewestPodcastsHandler().Handle),
		"/getPlaylists":             util.AsHandlerFunc(handlers.NewGetPlaylistsHandler(appCtx.SubsonicService).Handle),
		"/getPlaylist":              util.AsHandlerFunc(handlers.NewGetPlaylistHandler(appCtx.SubsonicService).Handle),
		"/getPodcasts":              util.AsHandlerFunc(handlers.NewGetPodcastsHandler().Handle),
		"/getRandomSongs":           util.AsHandlerFunc(handlers.NewGetRandomSongsHandler(appCtx.SubsonicService).Handle),
		"/getScanStatus":            util.AsHandlerFunc(handlers.NewGetScanStatusHandler().Handle),
		"/getSong":                  util.AsHandlerFunc(handlers.NewGetSongHandler(appCtx.SubsonicService).Handle),
		"/getStarred2":              util.AsHandlerFunc(handlers.NewGetStarred2Handler().Handle),
		"/search3":                  util.AsHandlerFunc(handlers.NewSearch3Handler(appCtx.SubsonicService).Handle),

		"/scrobble": util.AsHandlerFunc(handlers.NewScrobbleHandler(appCtx.SubsonicService).Handle),

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
			util.LogWarning(r, fmt.Sprintf("Handler is not implemented for %s %s", r.Method, r.URL.Path))
		},
	)

	return resultHandlers
}
