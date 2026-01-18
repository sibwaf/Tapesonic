package listenbrainz

import (
	"fmt"
	"tapesonic/users"
	"tapesonic/util"
	"time"
)

type ListenBrainzService struct {
	sessions *ListenBrainzSessionStorage
	client   *ListenBrainzClient
}

func newListenBrainzService(sessions *ListenBrainzSessionStorage, client *ListenBrainzClient) *ListenBrainzService {
	return &ListenBrainzService{
		sessions: sessions,
		client:   client,
	}
}

func (s *ListenBrainzService) CreateSession(user users.User, token string) (ListenBrainzSession, error) {
	session, err := s.client.ValidateToken(token)
	if err != nil {
		return ListenBrainzSession{}, err
	}

	if !session.Valid {
		return ListenBrainzSession{}, fmt.Errorf("listenbrainz token is not valid")
	}

	return s.sessions.Save(ListenBrainzSession{
		UserId: user.Id,

		Token:    token,
		Username: session.Username,

		IsScrobblingEnabled: true,

		CreatedAt: util.NewTimestampWrapper(time.Now()),
		UpdatedAt: util.NewTimestampWrapper(time.Now()),
	})
}

func (s *ListenBrainzService) UpdateSessionSettings(user users.User, settings SessionSettings) (ListenBrainzSession, error) {
	return s.sessions.UpdateSettings(user.Id, settings)
}

func (s *ListenBrainzService) DeleteSession(user users.User) error {
	return s.sessions.Delete(user.Id)
}

func (s *ListenBrainzService) GetCurrentSession(user users.User) (*ListenBrainzSession, error) {
	session, err := s.sessions.Find(user.Id)
	if err != nil {
		return &ListenBrainzSession{}, err
	} else if session == nil {
		return nil, nil
	}

	return session, err
}

func (s *ListenBrainzService) UpdateNowPlaying(user users.User, artist string, title string, album string) error {
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

	return s.client.SubmitListens(
		session.Token,
		SubmitListensRequest{
			ListenType: ListenTypePlayingNow,
			Payload: []SubmitListensRequestPayloadItem{
				{
					TrackMetadata: SubmitListensRequestPayloadItemTrackMetadata{
						ArtistName:  artist,
						ReleaseName: album,
						TrackName:   title,
					},
				},
			},
		},
	)
}

func (s *ListenBrainzService) Scrobble(user users.User, timestamp time.Time, artist string, title string, album string) error {
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

	return s.client.SubmitListens(
		session.Token,
		SubmitListensRequest{
			ListenType: ListenTypeSingle,
			Payload: []SubmitListensRequestPayloadItem{
				{
					ListenedAt: timestamp.Unix(),
					TrackMetadata: SubmitListensRequestPayloadItemTrackMetadata{
						ArtistName:  artist,
						ReleaseName: album,
						TrackName:   title,
					},
				},
			},
		},
	)
}
