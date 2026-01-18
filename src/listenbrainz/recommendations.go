package listenbrainz

import (
	"strings"

	"github.com/google/uuid"
)

const (
	playlistFetchLimit = 20
)

type ListenBrainzRecommendationService struct {
	sessions *ListenBrainzSessionStorage
	client   *ListenBrainzClient
}

func newListenBrainzRecommendationService(
	sessions *ListenBrainzSessionStorage,
	client *ListenBrainzClient,
) *ListenBrainzRecommendationService {
	return &ListenBrainzRecommendationService{
		sessions: sessions,
		client:   client,
	}
}

func (svc *ListenBrainzRecommendationService) GetAllSessions() ([]ListenBrainzSession, error) {
	return svc.sessions.GetAllSessions()
}

func (svc *ListenBrainzRecommendationService) GetPlaylists(userId uuid.UUID) ([]PlaylistResponse, error) {
	session, err := svc.sessions.Find(userId)
	if err != nil {
		return []PlaylistResponse{}, err
	}
	if session == nil {
		return []PlaylistResponse{}, ErrNoActiveSession
	}

	result := []PlaylistResponse{}
	offset := 0
	for {
		playlists, err := svc.client.GetPlaylistsCreatedFor(session.Token, session.Username, playlistFetchLimit, offset)
		if err != nil {
			return []PlaylistResponse{}, err
		}

		for _, wrapper := range playlists.Playlists {
			playlist := wrapper.Playlist
			playlist.Identifier = normalizeIdentifier(playlist.Identifier)

			result = append(result, playlist)
		}

		offset += len(playlists.Playlists)
		if len(playlists.Playlists) < playlistFetchLimit {
			break
		}
	}

	return result, nil
}

func (svc *ListenBrainzRecommendationService) GetPlaylist(userId uuid.UUID, playlistId string) (PlaylistResponse, error) {
	session, err := svc.sessions.Find(userId)
	if err != nil {
		return PlaylistResponse{}, err
	}
	if session == nil {
		return PlaylistResponse{}, ErrNoActiveSession
	}

	playlist, err := svc.client.GetPlaylist(session.Token, playlistId)
	if err != nil {
		return PlaylistResponse{}, err
	}

	playlist.Identifier = normalizeIdentifier(playlist.Identifier)

	return playlist, nil
}

func normalizeIdentifier(identifier string) string {
	identifierParts := strings.Split(identifier, "/")
	return identifierParts[len(identifierParts)-1]
}
