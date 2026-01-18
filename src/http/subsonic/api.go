package subsonic

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"tapesonic/appcontext"
	"tapesonic/http/subsonic/util"
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/users"
)

type SubsonicRawHandler func(w http.ResponseWriter, r *http.Request) (response *subsonic.Response, err error)

func asRawHandlerFunc(handler SubsonicRawHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(w, r)
		if err != nil {
			if errors.Is(err, context.Canceled) && errors.Is(r.Context().Err(), context.Canceled) {
				util.LogDebug(r, "Client cancelled the request")
			} else if errors.Is(err, model.ErrNotAuthenticated) {
				util.LogDebug(r, "Not authenticated")
				response = subsonic.NewFailedResponse(subsonic.ERROR_CODE_NOT_AUTHENTICATED, "Not authenticated")
			} else if errors.Is(err, model.ErrMissingParameter) {
				util.LogDebug(r, err.Error())
				response = subsonic.NewFailedResponse(subsonic.ERROR_CODE_PARAMETER_MISSING, err.Error())
			} else {
				util.LogError(r, fmt.Sprintf("Failed to process request: %s", err.Error()))
				response = subsonic.NewFailedResponse(subsonic.ERROR_CODE_GENERIC, "Server failed to process the request")
			}
		}

		if response != nil {
			writeResponse(w, r, response)
		}
	}
}

type SubsonicHandler func(r *http.Request) (*subsonic.Response, error)

func asHandlerFunc(handler SubsonicHandler) http.HandlerFunc {
	return asRawHandlerFunc(func(w http.ResponseWriter, r *http.Request) (response *subsonic.Response, err error) {
		return handler(r)
	})
}

func writeResponse(w http.ResponseWriter, r *http.Request, response *subsonic.Response) {
	wrappedResponse := subsonic.ResponseWrapper{
		SubsonicResponse: *response,
	}

	switch format := getFormat(r); format {
	case subsonic.FORMAT_JSON:
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wrappedResponse)
	case subsonic.FORMAT_XML, "":
		w.Header().Add("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(wrappedResponse)
	default:
		util.LogError(r, "Unsupported format", "format", format)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getFormat(r *http.Request) string {
	return r.URL.Query().Get(subsonic.QUERY_FORMAT)
}

type authenticator struct {
	users *users.UserService
}

func newAuthenticator(users *users.UserService) *authenticator {
	return &authenticator{users: users}
}

func (auth *authenticator) Authenticate(r *http.Request) (users.User, error) {
	query := r.URL.Query()

	username := query.Get(subsonic.QUERY_USERNAME)
	if username == "" {
		return users.User{}, model.NewErrMissingParameter(subsonic.QUERY_USERNAME)
	}

	password := query.Get(subsonic.QUERY_PASSWORD)

	token := query.Get(subsonic.QUERY_TOKEN)
	salt := query.Get(subsonic.QUERY_SALT)

	if token != "" || salt != "" {
		if salt == "" {
			return users.User{}, model.NewErrMissingParameter(subsonic.QUERY_SALT)
		}
		if token == "" {
			return users.User{}, model.NewErrMissingParameter(subsonic.QUERY_TOKEN)
		}

		user, err := auth.users.FindByName(username)
		if err != nil {
			return users.User{}, err
		}
		if user == nil {
			return users.User{}, model.ErrNotAuthenticated
		}

		expectedToken := subsonic.GenerateAuthToken(user.ApiKey, salt)
		if token == expectedToken {
			return *user, nil
		} else {
			return users.User{}, model.ErrNotAuthenticated
		}
	} else if password != "" {
		passwordVariants := []string{password}

		if strings.HasPrefix(password, "enc:") {
			decodedPassword, err := subsonic.DecodePassword(password)
			if err == nil {
				passwordVariants = append(passwordVariants, decodedPassword)
			}
		}

		for _, passwordVariant := range passwordVariants {
			user, err := auth.users.TryAuthenticateWithPassword(username, passwordVariant)
			if err != nil {
				return users.User{}, err
			} else if user != nil {
				return *user, nil
			}

			user, err = auth.users.TryAuthenticateWithApiKey(passwordVariant)
			if err != nil {
				return users.User{}, err
			} else if user != nil && user.Name == username {
				return *user, nil
			}
		}

		return users.User{}, model.ErrNotAuthenticated
	} else {
		return users.User{}, model.NewErrMissingParameter(subsonic.QUERY_PASSWORD)
	}
}

func GetHandlers(appCtx *appcontext.Context) map[string]http.HandlerFunc {
	auth := newAuthenticator(appCtx.Users.UserService)

	rawHandlers := map[string]http.HandlerFunc{
		"/ping": asHandlerFunc(Ping(auth)),

		"/getAlbumList2":            asHandlerFunc(GetAlbumList2(auth, appCtx.Library.LibraryService)),
		"/getAlbum":                 asHandlerFunc(GetAlbum(auth, appCtx.Library.LibraryService)),
		"/getArtists":               asHandlerFunc(GetArtists(auth)),
		"/getArtist":                asHandlerFunc(GetArtist(auth, appCtx.Library.LibraryService)),
		"/getGenres":                asHandlerFunc(GetGenres(auth)),
		"/getInternetRadioStations": asHandlerFunc(GetInternetRadioStations(auth)),
		"/getLicense":               asHandlerFunc(GetLicense(auth)),
		"/getMusicFolders":          asHandlerFunc(GetMusicFolders(auth)),
		"/getPodcasts":              asHandlerFunc(GetPodcasts(auth)),
		"/getNewestPodcasts":        asHandlerFunc(GetNewestPodcasts(auth)),
		"/getPlaylists":             asHandlerFunc(GetPlaylists(auth, appCtx.Library.LibraryService)),
		"/getPlaylist":              asHandlerFunc(GetPlaylist(auth, appCtx.Library.LibraryService)),
		"/getRandomSongs":           asHandlerFunc(GetRandomSongs(auth, appCtx.Library.LibraryService)),
		"/getScanStatus":            asHandlerFunc(GetScanStatus(auth)),
		"/getSong":                  asHandlerFunc(GetSong(auth, appCtx.Library.LibraryService)),
		"/getStarred2":              asHandlerFunc(GetStarred2(auth)),
		"/search3":                  asHandlerFunc(Search3(auth, appCtx.Library.LibraryService)),

		"/scrobble": asHandlerFunc(Scrobble(auth, appCtx.Scrobbling.ScrobbleService)),

		"/stream":      asRawHandlerFunc(Stream(auth, appCtx.Library.LibraryService, appCtx.Media.Streams)),
		"/getCoverArt": asRawHandlerFunc(GetCoverArt(auth, appCtx.Library.LibraryService, appCtx.Media.Covers)),
	}

	resultHandlers := map[string]http.HandlerFunc{}
	for path, handler := range rawHandlers {
		wrappedHandler := util.Logged(handler)
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
