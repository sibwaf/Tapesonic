package media

import (
	"errors"
	"log/slog"
	"net/http"
	"tapesonic/appcontext"
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

type MediaHandler func(r *http.Request, w http.ResponseWriter) error

func asHandlerFunc(handler MediaHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r, w)
		if err != nil {
			handleError(err, w)
			return
		}
	}
}

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	auth := newAuthenticator(appCtx.Users.UserService)
	router := mux.NewRouter()

	router.Handle("/media/thumbnails/{thumbnailId}", asHandlerFunc(GetThumbnail(auth, appCtx.Library.LibraryService, appCtx.Media.Covers))).Methods("GET")

	return map[string]http.HandlerFunc{"/media/": router.ServeHTTP}
}
