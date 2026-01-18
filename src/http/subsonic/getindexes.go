package subsonic

import (
	"net/http"
	"tapesonic/subsonic"
	"time"
)

func GetIndexes(auth *authenticator) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: implement

		response := subsonic.NewOkResponse()
		response.Indexes = &subsonic.Indexes{
			IgnoredArticles: "",
			LastModified:    time.Now().UnixMilli(),
			Shortcut:        []subsonic.ArtistId3{},
			Child:           []subsonic.Child{},
			Index:           []subsonic.IndexId3{},
		}
		return response, nil
	}
}
