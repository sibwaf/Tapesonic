package admin

import (
	"encoding/json"
	"net/http"
	"tapesonic/lastfm"
	"tapesonic/model"
	"time"
)

type LastFmAuthLinkRs struct {
	Url   string
	Token string
}

func PostLastFmAuthLink(auth *authenticator, lastfm *lastfm.LastFmService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		link, err := lastfm.CreateAuthLink()
		if err != nil {
			return nil, err
		}

		return LastFmAuthLinkRs{
			Url:   link.Url,
			Token: link.Token,
		}, nil
	}
}

type LastFmSessionRs struct {
	Username            string
	IsScrobblingEnabled bool
	UpdatedAt           time.Time
}

func toLastFmSessionRs(session lastfm.LastFmSession) LastFmSessionRs {
	return LastFmSessionRs{
		Username:            session.Username,
		IsScrobblingEnabled: session.IsScrobblingEnabled,
		UpdatedAt:           session.UpdatedAt.Unwrap(),
	}
}

func GetLastFmSession(auth *authenticator, lastfm *lastfm.LastFmService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		session, err := lastfm.GetCurrentSession(user)
		if err != nil {
			return nil, err
		}

		if session == nil {
			return nil, model.ErrNotFound
		}

		return toLastFmSessionRs(*session), nil
	}
}

type CreateLastFmSessionRq struct {
	Token string
}

func PostLastFmSession(auth *authenticator, lastfm *lastfm.LastFmService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		request := CreateLastFmSessionRq{}
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return nil, err
		}

		session, err := lastfm.CreateSession(user, request.Token)
		if err != nil {
			return nil, err
		}

		return toLastFmSessionRs(session), nil
	}
}

func DeleteLastFmSession(auth *authenticator, lastfm *lastfm.LastFmService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		return nil, lastfm.DeleteSession(user)
	}
}

type LastFmSessionSettingsRq struct {
	IsScrobblingEnabled bool
}

func PutLastFmSessionSettings(auth *authenticator, lastfmSvc *lastfm.LastFmService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		var rq LastFmSessionSettingsRq
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		session, err := lastfmSvc.UpdateSessionSettings(
			user,
			lastfm.SessionSettings{IsScrobblingEnabled: rq.IsScrobblingEnabled},
		)
		if err != nil {
			return nil, err
		}

		return toLastFmSessionRs(session), nil
	}
}
