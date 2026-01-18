package recommendations

import (
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/lastfm"
	"tapesonic/model"
	"tapesonic/search"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type LastFmRecommendedPlaylistService struct {
	lastfm    *lastfm.LastFmRecommendationService
	search    *search.SearchService
	playlists *RecommendationStorage
	planner   *TaskPlanner
}

func newLastFmRecommendedPlaylistService(
	lastfm *lastfm.LastFmRecommendationService,
	search *search.SearchService,
	playlists *RecommendationStorage,
	planner *TaskPlanner,
) *LastFmRecommendedPlaylistService {
	return &LastFmRecommendedPlaylistService{
		lastfm:    lastfm,
		search:    search,
		playlists: playlists,
		planner:   planner,
	}
}

func (svc *LastFmRecommendedPlaylistService) DiscoverSessions() error {
	// todo: this seems stupid,
	//   but i can't be bothered to waste any more time on this

	sessions, err := svc.lastfm.GetAllSessions()
	if err != nil {
		return err
	}

	for _, session := range sessions {
		err := svc.planner.ScheduleLastFmPlaylistDiscovery(session.UserId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *LastFmRecommendedPlaylistService) DiscoverPlaylists(userId uuid.UUID) error {
	slog.Info(fmt.Sprintf("Discovering last.fm recommended playlists for user id=%s", userId))

	syncTag := uuid.New().String()

	playlists, err := svc.lastfm.GetPlaylists(userId)
	if errors.Is(err, lastfm.ErrNoActiveSession) {
		slog.Warn(fmt.Sprintf("Cancelling last.fm recommended playlist discovery for missing session userId=%s", userId))
		return svc.planner.CancelLastFmPlaylistDiscovery(userId)
	} else if err != nil {
		return err
	}

	for _, rawPlaylist := range playlists {
		playlist := RecommendedPlaylist{
			Id:                 uuid.New(),
			Provider:           RecommendationProviderLastFm,
			ProviderPlaylistId: rawPlaylist.Id,
			UserId:             userId,
			Name:               rawPlaylist.Name,
			CreatedAt:          util.NewTimestampWrapper(time.Now()),
			UpdatedAt:          util.NewTimestampWrapper(time.Now()),
			SyncTag:            syncTag,
		}

		playlist, err := svc.playlists.UpsertPlaylist(playlist)
		if err != nil {
			return err
		}

		err = svc.planner.ScheduleLastFmPlaylistSync(playlist.Id)
		if err != nil {
			return err
		}
	}

	return svc.playlists.DeleteUnsyncedPlaylists(RecommendationProviderLastFm, userId, syncTag)
}

func (svc *LastFmRecommendedPlaylistService) RefreshPlaylist(playlistId uuid.UUID) error {
	slog.Info(fmt.Sprintf("Syncing last.fm recommended playlist id=%s", playlistId))

	playlist, err := svc.playlists.GetById(playlistId)
	if errors.Is(err, model.ErrNotFound) {
		slog.Warn(fmt.Sprintf("Cancelling last.fm recommended playlist sync for non-existing playlist id=%s", playlistId))
		return svc.planner.CancelLastFmPlaylistSync(playlistId)
	} else if err != nil {
		return err
	}

	targetSize := 50 // todo: config
	tracks, err := svc.collectTracks(playlist.UserId, playlist.Id, playlist.ProviderPlaylistId, targetSize)
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

func (svc *LastFmRecommendedPlaylistService) collectTracks(userId uuid.UUID, playlistId uuid.UUID, rawPlaylistId string, targetSize int) ([]RecommendedPlaylistTrack, error) {
	result := []RecommendedPlaylistTrack{}

	seenUrls := util.NewCountingSet[string]()
	for len(result) < targetSize && seenUrls.TotalSize() < targetSize*3 {
		playlistTracks, err := svc.lastfm.GetMorePlaylistTracks(userId, rawPlaylistId)
		if err != nil {
			return result, err
		}

		if len(playlistTracks) == 0 {
			break
		}

		for _, item := range playlistTracks {
			if seenUrls.Add(item.Url) > 1 {
				// skip duplicates
				continue
			}

			trackForSearch := search.TrackForSearch{
				Artist: item.Artists[0].Name,
				Title:  item.Name,
				Urls:   []string{},
			}
			for _, playlink := range item.Playlinks {
				trackForSearch.Urls = append(trackForSearch.Urls, playlink.Url)
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

			if len(result) >= targetSize {
				break
			}
		}
	}

	slog.Debug(fmt.Sprintf("Found %d/%d tracks for playlist id=%s", len(result), targetSize, playlistId))

	return result, nil
}
