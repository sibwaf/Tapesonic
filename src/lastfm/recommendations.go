package lastfm

import (
	"fmt"
	"tapesonic/model"

	"github.com/google/uuid"
)

type LastFmRecommendationService struct {
	sessions *LastFmSessionStorage
	client   *LastFmClient
}

func newLastFmRecommendationService(
	sessions *LastFmSessionStorage,
	client *LastFmClient,
) *LastFmRecommendationService {
	return &LastFmRecommendationService{
		sessions: sessions,
		client:   client,
	}
}

func (svc *LastFmRecommendationService) GetAllSessions() ([]LastFmSession, error) {
	return svc.sessions.GetAllSessions()
}

func (svc *LastFmRecommendationService) GetPlaylists(userId uuid.UUID) ([]Playlist, error) {
	session, err := svc.sessions.Find(userId)
	if err != nil {
		return []Playlist{}, err
	}
	if session == nil {
		return []Playlist{}, ErrNoActiveSession
	}

	return []Playlist{
		{Id: fmt.Sprintf("%s-library", userId.String()), Name: "last.fm: Library"},
		{Id: fmt.Sprintf("%s-mix", userId.String()), Name: "last.fm: Mix"},
		{Id: fmt.Sprintf("%s-recommended", userId.String()), Name: "last.fm: Recommended"},
	}, nil
}

func (svc *LastFmRecommendationService) GetMorePlaylistTracks(userId uuid.UUID, playlistId string) ([]PlaylistItemRs, error) {
	// last.fm returns different results for the same page each time

	session, err := svc.sessions.Find(userId)
	if err != nil {
		return []PlaylistItemRs{}, err
	}
	if session == nil {
		return []PlaylistItemRs{}, ErrNoActiveSession
	}

	var playlistWrapper PlaylistWrapperRs
	if playlistId == fmt.Sprintf("%s-library", userId.String()) {
		playlistWrapper, err = svc.client.GetLibraryPlaylist(session.Username, 1)
	} else if playlistId == fmt.Sprintf("%s-mix", userId.String()) {
		playlistWrapper, err = svc.client.GetMixPlaylist(session.Username, 1)
	} else if playlistId == fmt.Sprintf("%s-recommended", userId.String()) {
		playlistWrapper, err = svc.client.GetRecommendedPlaylist(session.Username, 1)
	} else {
		return []PlaylistItemRs{}, model.ErrNotFound
	}

	if err != nil {
		return []PlaylistItemRs{}, err
	}

	return playlistWrapper.Items, err
}
