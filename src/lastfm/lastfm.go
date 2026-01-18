package lastfm

import (
	"fmt"
	"tapesonic/users"
	"tapesonic/util"
	"time"
)

type LastFmService struct {
	client   *LastFmClient
	sessions *LastFmSessionStorage
}

func newLastFmService(
	client *LastFmClient,
	sessions *LastFmSessionStorage,
) *LastFmService {
	return &LastFmService{
		client:   client,
		sessions: sessions,
	}
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

func (s *LastFmService) CreateSession(user users.User, token string) (LastFmSession, error) {
	if s.client == nil {
		return LastFmSession{}, ErrLastFmNotConfigured
	}

	session, err := s.client.AuthGetSession(token)
	if err != nil {
		return LastFmSession{}, err
	}

	return s.sessions.Save(LastFmSession{
		UserId: user.Id,

		SessionKey: session.Session.Key,
		Username:   session.Session.Name,

		IsScrobblingEnabled: true,

		CreatedAt: util.NewTimestampWrapper(time.Now()),
		UpdatedAt: util.NewTimestampWrapper(time.Now()),
	})
}

func (s *LastFmService) UpdateSessionSettings(user users.User, settings SessionSettings) (LastFmSession, error) {
	return s.sessions.UpdateSettings(user.Id, settings)
}

func (s *LastFmService) DeleteSession(user users.User) error {
	return s.sessions.Delete(user.Id)
}

func (s *LastFmService) GetCurrentSession(user users.User) (*LastFmSession, error) {
	session, err := s.sessions.Find(user.Id)
	if err != nil {
		return &LastFmSession{}, err
	} else if session == nil {
		return nil, nil
	}

	return session, err
}

func (s *LastFmService) UpdateNowPlaying(user users.User, artist string, title string, album string) error {
	if s.client == nil {
		return ErrLastFmNotConfigured
	}

	session, err := s.sessions.Find(user.Id)
	if err != nil {
		return err
	} else if session == nil {
		return ErrNoActiveSession
	}

	if !session.IsScrobblingEnabled {
		return nil
	}

	if artist == "" || title == "" {
		return fmt.Errorf("artist and title must be provided")
	}

	return s.client.UpdateNowPlaying(session.SessionKey, UpdateNowPlayingRq{Artist: artist, Track: title, Album: album})
}

func (s *LastFmService) Scrobble(user users.User, timestamp time.Time, artist string, title string, album string) error {
	if s.client == nil {
		return ErrLastFmNotConfigured
	}

	session, err := s.sessions.Find(user.Id)
	if err != nil {
		return err
	} else if session == nil {
		return ErrNoActiveSession
	}

	if !session.IsScrobblingEnabled {
		return nil
	}

	if artist == "" || title == "" {
		return fmt.Errorf("artist and title must be provided")
	}

	return s.client.Scrobble(
		session.SessionKey,
		ScrobbleRq{
			Timestamp: timestamp.Unix(),
			Artist:    artist,
			Track:     title,
			Album:     album,
		},
	)
}
