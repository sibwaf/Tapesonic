package admin

import (
	"encoding/json"
	"net/http"
	"tapesonic/listenbrainz"
	"tapesonic/model"
	"time"
)

type ListenBrainzSessionRs struct {
	Username            string
	IsScrobblingEnabled bool
	UpdatedAt           time.Time
}

func toListenBrainzSessionRs(session listenbrainz.ListenBrainzSession) ListenBrainzSessionRs {
	return ListenBrainzSessionRs{
		Username:            session.Username,
		IsScrobblingEnabled: session.IsScrobblingEnabled,
		UpdatedAt:           session.UpdatedAt.Unwrap(),
	}
}

func GetListenBrainzSession(auth *authenticator, listenbrainz *listenbrainz.ListenBrainzService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		session, err := listenbrainz.GetCurrentSession(user)
		if err != nil {
			return nil, err
		}

		if session == nil {
			return nil, model.ErrNotFound
		}

		return toListenBrainzSessionRs(*session), nil
	}
}

type CreateListenBrainzSessionRq struct {
	Token string
}

func PostListenBrainzSession(auth *authenticator, listenbrainz *listenbrainz.ListenBrainzService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		request := CreateListenBrainzSessionRq{}
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return nil, err
		}

		session, err := listenbrainz.CreateSession(user, request.Token)
		if err != nil {
			return nil, err
		}

		return toListenBrainzSessionRs(session), nil
	}
}

func DeleteListenBrainzSession(auth *authenticator, listenbrainz *listenbrainz.ListenBrainzService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		return nil, listenbrainz.DeleteSession(user)
	}
}

type ListenBrainzSessionSettingsRq struct {
	IsScrobblingEnabled bool
}

func PutListenBrainzSessionSettings(auth *authenticator, listenbrainzSvc *listenbrainz.ListenBrainzService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var rq ListenBrainzSessionSettingsRq
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		session, err := listenbrainzSvc.UpdateSessionSettings(
			user,
			listenbrainz.SessionSettings{IsScrobblingEnabled: rq.IsScrobblingEnabled},
		)
		if err != nil {
			return nil, err
		}

		return toListenBrainzSessionRs(session), nil
	}
}
