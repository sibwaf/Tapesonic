package logic

import (
	"errors"
	"fmt"
	"tapesonic/http/lastfm"
	"tapesonic/storage"
	"time"
)

var (
	ErrLastFmNotConfigured = errors.New("last.fm client is not configured")
)

type LastFmService struct {
	client   *lastfm.LastFmClient
	sessions *storage.LastFmSessionStorage
}

func NewLastFmService(
	client *lastfm.LastFmClient,
	sessions *storage.LastFmSessionStorage,
) *LastFmService {
	return &LastFmService{
		client:   client,
		sessions: sessions,
	}
}

type LastFmAuthLink struct {
	Url   string
	Token string
}

func (s *LastFmService) CreateAuthLink() (LastFmAuthLink, error) {
	if s.client == nil {
		return LastFmAuthLink{}, ErrLastFmNotConfigured
	}

	token, err := s.client.AuthGetToken()
	if err != nil {
		return LastFmAuthLink{}, err
	}

	return LastFmAuthLink{
		Url:   fmt.Sprintf("http://www.last.fm/api/auth/?api_key=%s&token=%s", s.client.GetApiKey(), token.Token),
		Token: token.Token,
	}, nil
}

type LastFmSession struct {
	Username   string
	SessionKey string
	UpdatedAt  time.Time
}

func (s *LastFmService) CreateSession(token string) (LastFmSession, error) {
	if s.client == nil {
		return LastFmSession{}, ErrLastFmNotConfigured
	}

	session, err := s.client.AuthGetSession(token)
	if err != nil {
		return LastFmSession{}, err
	}

	savedSession, err := s.sessions.Save(storage.LastFmSession{
		SessionKey: session.Session.Key,
		Username:   session.Session.Name,
	})
	if err != nil {
		return LastFmSession{}, err
	}

	return LastFmSession{
		Username:   savedSession.Username,
		SessionKey: savedSession.SessionKey,
		UpdatedAt:  savedSession.UpdatedAt,
	}, nil
}

func (s *LastFmService) GetCurrentSession() (*LastFmSession, error) {
	session, err := s.sessions.Find()
	if err != nil {
		return &LastFmSession{}, err
	} else if session == nil {
		return nil, nil
	}

	return &LastFmSession{
		Username:   session.Username,
		SessionKey: session.SessionKey,
		UpdatedAt:  session.UpdatedAt,
	}, err
}
