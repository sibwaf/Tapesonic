package admin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"tapesonic/appcontext"
	"tapesonic/http/admin/handlers"
	"tapesonic/model"
	"tapesonic/users"

	"github.com/gorilla/mux"
)

func handleError(err error, w http.ResponseWriter) {
	if errors.Is(err, model.ErrNotAuthenticated) {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"master\", charset=\"UTF-8\"")
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if errors.Is(err, model.ErrInsufficientPermissions) {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if errors.Is(err, model.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		// todo
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(err.Error())
		return
	}
}

type WebappRawHandler func(r *http.Request, w http.ResponseWriter) error

func rawAsHandlerFunc(handler WebappRawHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r, w)
		if err != nil {
			handleError(err, w)
			return
		}
	}
}

type WebappHandler func(r *http.Request) (any, error)

func asHandlerFunc(handler WebappHandler) http.HandlerFunc {
	return rawAsHandlerFunc(func(r *http.Request, w http.ResponseWriter) error {
		response, err := handler(r)
		if err != nil {
			return err
		}

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(response)

		return nil
	})
}

type authenticator struct {
	users *users.UserService
}

func newAuthenticator(users *users.UserService) *authenticator {
	return &authenticator{users: users}
}

func (auth *authenticator) Authenticate(r *http.Request) (users.User, error) {
	username, password, authSucceeded := r.BasicAuth()
	if !authSucceeded {
		return users.User{}, model.ErrNotAuthenticated
	}

	user, err := auth.users.TryAuthenticateWithPassword(username, password)
	if err != nil {
		return users.User{}, err
	}
	if user == nil {
		return users.User{}, model.ErrNotAuthenticated
	}

	return *user, nil
}

func (auth *authenticator) Authorize(r *http.Request, role model.Role) (users.User, error) {
	user, err := auth.Authenticate(r)
	if err != nil {
		return user, err
	}

	if user.Role != role {
		return user, model.ErrInsufficientPermissions
	}

	return user, nil
}

func (auth *authenticator) Authenticated(handler WebappHandler) WebappHandler {
	// todo: deprecated

	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		return handler(r)
	}
}

func (auth *authenticator) AuthenticatedRaw(handler WebappRawHandler) WebappRawHandler {
	// todo: deprecated

	return func(r *http.Request, w http.ResponseWriter) error {
		_, err := auth.Authenticate(r)
		if err != nil {
			return err
		}

		return handler(r, w)
	}
}

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	auth := newAuthenticator(appCtx.Users.UserService)

	type PathHandler struct {
		Path    string
		Handler WebappHandler
	}

	// todo: logging
	rawHandlers := []PathHandler{
		{Path: "/api/thumbnails", Handler: handlers.NewThumbnailsHandler(appCtx.ThumbnailService).Handle},
	}

	router := mux.NewRouter()
	for _, pathHandler := range rawHandlers {
		router.HandleFunc(pathHandler.Path, asHandlerFunc(auth.Authenticated(pathHandler.Handler)))
	}

	router.HandleFunc("/media/thumbnails/{thumbnailId}", rawAsHandlerFunc(auth.AuthenticatedRaw(handlers.NewThumbnailRawHandler(appCtx.ThumbnailService).Handle)))

	router.HandleFunc("/api/users", asHandlerFunc(GetUsers(auth, appCtx.Users.UserService))).Methods("GET")
	router.HandleFunc("/api/users", asHandlerFunc(PostUsers(auth, appCtx.Users.UserService))).Methods("POST")
	router.HandleFunc("/api/users/me", asHandlerFunc(GetUserMe(auth, appCtx.Users.UserService))).Methods("GET")
	router.HandleFunc("/api/users/root", asHandlerFunc(PutUserRoot(appCtx.Users.UserService))).Methods("PUT")
	router.HandleFunc("/api/users/{userId}", asHandlerFunc(PatchUser(auth, appCtx.Users.UserService))).Methods("PATCH")
	router.HandleFunc("/api/users/{userId}/api-keys", asHandlerFunc(PostUserApiKeys(auth, appCtx.Users.UserService))).Methods("POST")

	router.HandleFunc("/api/sources", asHandlerFunc(GetSources(auth, appCtx.Sources.SourceService))).Methods("GET")
	router.HandleFunc("/api/sources/{sourceId}", asHandlerFunc(GetSource(auth, appCtx.Sources.SourceService))).Methods("GET")
	router.HandleFunc("/api/sources/{sourceId}/hierarchy", asHandlerFunc(GetSourceHierarchy(auth, appCtx.Sources.SourceService))).Methods("GET")
	router.HandleFunc("/api/sources/{sourceId}/tracks", asHandlerFunc(GetSourceTracks(auth, appCtx.Sources.SourceService))).Methods("GET")
	router.HandleFunc("/api/sources/{sourceId}/tracks", asHandlerFunc(PutSourceTracks(auth, appCtx.Sources.SourceService))).Methods("PUT")
	router.HandleFunc("/api/sources/{sourceId}/file", asHandlerFunc(DeleteSourceFile(auth, appCtx.Sources.SourceService))).Methods("DELETE")

	router.HandleFunc("/api/remotes", asHandlerFunc(GetRemotes(auth, appCtx.Remotes.RemoteService))).Methods("GET")
	router.HandleFunc("/api/remotes", asHandlerFunc(PostRemotes(auth, appCtx.Remotes.RemoteService))).Methods("POST")
	router.HandleFunc("/api/remotes/{remoteId}", asHandlerFunc(GetRemote(auth, appCtx.Remotes.RemoteService))).Methods("GET")
	router.HandleFunc("/api/remotes/{remoteId}", asHandlerFunc(PutRemote(auth, appCtx.Remotes.RemoteService))).Methods("PUT")
	router.HandleFunc("/api/remotes/{remoteId}", asHandlerFunc(DeleteRemote(auth, appCtx.Remotes.RemoteService))).Methods("DELETE")
	router.HandleFunc("/api/remotes/{remoteId}/auth", asHandlerFunc(PutRemoteAuth(auth, appCtx.Remotes.RemoteService))).Methods("PUT")
	router.HandleFunc("/api/remotes/{remoteId}/auth", asHandlerFunc(DeleteRemoteAuth(auth, appCtx.Remotes.RemoteService))).Methods("DELETE")

	router.HandleFunc("/api/listenbrainz/auth", asHandlerFunc(GetListenBrainzSession(auth, appCtx.ListenBrainz.ListenBrainzService))).Methods("GET")
	router.HandleFunc("/api/listenbrainz/auth", asHandlerFunc(PostListenBrainzSession(auth, appCtx.ListenBrainz.ListenBrainzService))).Methods("POST")
	router.HandleFunc("/api/listenbrainz/auth", asHandlerFunc(DeleteListenBrainzSession(auth, appCtx.ListenBrainz.ListenBrainzService))).Methods("DELETE")
	router.HandleFunc("/api/listenbrainz/settings", asHandlerFunc(PutListenBrainzSessionSettings(auth, appCtx.ListenBrainz.ListenBrainzService))).Methods("PUT")

	router.HandleFunc("/api/lastfm/create-auth-link", asHandlerFunc(PostLastFmAuthLink(auth, appCtx.LastFm.LastFmService))).Methods("POST")
	router.HandleFunc("/api/lastfm/auth", asHandlerFunc(GetLastFmSession(auth, appCtx.LastFm.LastFmService))).Methods("GET")
	router.HandleFunc("/api/lastfm/auth", asHandlerFunc(PostLastFmSession(auth, appCtx.LastFm.LastFmService))).Methods("POST")
	router.HandleFunc("/api/lastfm/auth", asHandlerFunc(DeleteLastFmSession(auth, appCtx.LastFm.LastFmService))).Methods("DELETE")
	router.HandleFunc("/api/lastfm/settings", asHandlerFunc(PutLastFmSessionSettings(auth, appCtx.LastFm.LastFmService))).Methods("PUT")

	router.HandleFunc("/api/tracks", asHandlerFunc(GetTracks(auth, appCtx.Search.SearchService))).Methods("GET")

	router.HandleFunc("/api/tapes/guess-metadata", asHandlerFunc(PostTapesGuessMetadata(auth, appCtx.Tapes.TapeService))).Methods("POST")
	router.HandleFunc("/api/tapes", asHandlerFunc(GetTapes(auth, appCtx.Tapes.TapeService))).Methods("GET")
	router.HandleFunc("/api/tapes", asHandlerFunc(PostTapes(auth, appCtx.Tapes.TapeService))).Methods("POST")
	router.HandleFunc("/api/tapes/{tapeId}", asHandlerFunc(GetTape(auth, appCtx.Tapes.TapeService))).Methods("GET")
	router.HandleFunc("/api/tapes/{tapeId}", asHandlerFunc(PutTape(auth, appCtx.Tapes.TapeService))).Methods("PUT")
	router.HandleFunc("/api/tapes/{tapeId}", asHandlerFunc(DeleteTape(auth, appCtx.Tapes.TapeService))).Methods("DELETE")

	// todo: wow that's disgusting
	return map[string]http.HandlerFunc{
		"/api/":   router.ServeHTTP,
		"/media/": router.ServeHTTP,
	}
}
