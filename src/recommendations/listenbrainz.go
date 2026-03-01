package recommendations

import (
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/listenbrainz"
	"tapesonic/model"
	"tapesonic/search"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type ListenBrainzRecommendedPlaylistService struct {
	listenbrainz *listenbrainz.ListenBrainzRecommendationService
	search       *search.SearchService
	playlists    *RecommendationStorage
	planner      *TaskPlanner
}

func newListenBrainzRecommendedPlaylistService(
	listenbrainz *listenbrainz.ListenBrainzRecommendationService,
	search *search.SearchService,
	playlists *RecommendationStorage,
	planner *TaskPlanner,
) *ListenBrainzRecommendedPlaylistService {
	return &ListenBrainzRecommendedPlaylistService{
		listenbrainz: listenbrainz,
		search:       search,
		playlists:    playlists,
		planner:      planner,
	}
}

func (svc *ListenBrainzRecommendedPlaylistService) DiscoverSessions() error {
	// todo: this seems stupid,
	//   but i can't be bothered to waste any more time on this

	sessions, err := svc.listenbrainz.GetAllSessions()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		err := svc.planner.ScheduleListenBrainzPlaylistDiscovery(session.UserId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *ListenBrainzRecommendedPlaylistService) DiscoverPlaylists(userId uuid.UUID) error {
	slog.Info(fmt.Sprintf("Discovering ListenBrainz recommended playlists for user id=%s", userId))

	syncTag := uuid.New().String()

	playlists, err := svc.listenbrainz.GetPlaylists(userId)
	if errors.Is(err, listenbrainz.ErrNoActiveSession) {
		slog.Warn(fmt.Sprintf("Cancelling ListenBrainz recommended playlist discovery for missing session userId=%s", userId))
		return svc.planner.CancelListenBrainzPlaylistDiscovery(userId)
	} else if err != nil {
		return err
	}

	for _, rawPlaylist := range playlists {
		playlist := RecommendedPlaylist{
			Id:                 uuid.New(),
			Provider:           RecommendationProviderListenBrainz,
			ProviderPlaylistId: rawPlaylist.Identifier,
			UserId:             userId,
			Name:               rawPlaylist.Title,
			CreatedAt:          util.NewTimestampWrapper(rawPlaylist.Date),
			UpdatedAt:          util.NewTimestampWrapper(time.Now()),
			SyncTag:            syncTag,
		}

		playlist, err := svc.playlists.UpsertPlaylist(playlist)
		if err != nil {
			return err
		}

		err = svc.planner.ScheduleListenBrainzPlaylistSync(playlist.Id)
		if err != nil {
			return err
		}
	}

	return svc.playlists.DeleteUnsyncedPlaylists(RecommendationProviderListenBrainz, userId, syncTag)
}

func (svc *ListenBrainzRecommendedPlaylistService) RefreshPlaylist(playlistId uuid.UUID) error {
	slog.Info(fmt.Sprintf("Syncing ListenBrainz recommended playlist id=%s", playlistId))

	playlist, err := svc.playlists.GetById(playlistId)
	if errors.Is(err, model.ErrNotFound) {
		slog.Warn(fmt.Sprintf("Cancelling ListenBrainz recommended playlist sync for non-existing playlist id=%s", playlistId))
		return svc.planner.CancelListenBrainzPlaylistSync(playlistId)
	} else if err != nil {
		return err
	}

	tracks, err := svc.collectTracks(playlist.UserId, playlist.Id, playlist.ProviderPlaylistId)
	if err != nil {
		return err
	}

	err = svc.playlists.ReplaceTracks(playlist.Id, tracks)
	if err != nil {
		return err
	}

	// todo: decide what to do with createdAt/updatedAt
	err = svc.playlists.UpdatePlaylistUpdatedAt(playlist.Id, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (svc *ListenBrainzRecommendedPlaylistService) collectTracks(userId uuid.UUID, playlistId uuid.UUID, rawPlaylistId string) ([]RecommendedPlaylistTrack, error) {
	playlist, err := svc.listenbrainz.GetPlaylist(userId, rawPlaylistId)
	if err != nil {
		return []RecommendedPlaylistTrack{}, fmt.Errorf("failed to get playlist id=%s: %w", rawPlaylistId, err)
	}

	result := []RecommendedPlaylistTrack{}
	for _, item := range playlist.Track {
		trackForSearch := search.TrackForSearch{
			// todo: ArtistMusicBrainzId
			Artist: item.Creator,
			Title:  item.Title,
			Album:  item.Album,
		}

		foundTrack, err := svc.search.FindTrack(userId, trackForSearch)
		if err != nil {
			return []RecommendedPlaylistTrack{}, fmt.Errorf("track search failed: %w", err)
		}
		if foundTrack == nil {
			slog.Debug(fmt.Sprintf("Couldn't find a matching track for playlist id=%s: %+v", playlistId, trackForSearch))
			continue
		}

		result = append(
			result,
			RecommendedPlaylistTrack{
				Artist:  trackForSearch.Artist,
				Title:   trackForSearch.Title,
				TrackId: foundTrack.Id,
			},
		)
	}

	slog.Debug(fmt.Sprintf("Found %d/%d tracks for playlist id=%s", len(result), len(playlist.Track), playlistId))

	return result, nil
}
