package subsonic

import (
	"net/http"
	"tapesonic/model"
	"tapesonic/subsonic"
)

func GetUser(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			return subsonic.NewParameterMissingResponse("username"), nil
		}

		if username != user.Name && user.Role != model.ROLE_ADMIN {
			return subsonic.NewNotAuthorizedResponse(), nil
		}

		response := subsonic.NewOkResponse()
		response.User = &subsonic.User{
			Username:            user.Name,
			ScrobblingEnabled:   true,
			AdminRole:           user.Role == model.ROLE_ADMIN,
			SettingsRole:        false,
			DownloadRole:        false, // todo: download not implemented yet
			UploadRole:          false,
			PlaylistRole:        false, // todo: playlist api not implemented yet
			CoverArtRole:        true,
			CommentRole:         false,
			PodcastRole:         false, // todo: podcasts not implemented yet
			StreamRole:          true,
			JukeboxRole:         false,
			ShareRole:           false,
			VideoConversionRole: false,
		}
		return response, nil
	}
}
