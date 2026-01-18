package recommendations

import (
	"fmt"
	"log/slog"
	"tapesonic/lastfm"
	"tapesonic/listenbrainz"
	"tapesonic/scheduling"
	"tapesonic/search"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type RecommendationsModule struct {
	storage                          *RecommendationStorage
	lastFmRecommendedPlaylists       *LastFmRecommendedPlaylistService
	listenBrainzRecommendedPlaylists *ListenBrainzRecommendedPlaylistService
	playlistRefreshCron              cron.Schedule
}

func NewRecommendationModule(
	db *gorm.DB,
	lastFmRecommendations *lastfm.LastFmRecommendationService,
	listenBrainzRecommendations *listenbrainz.ListenBrainzRecommendationService,
	search *search.SearchService,
	taskScheduler *scheduling.TaskScheduler,
	playlistRefreshCron cron.Schedule,
) *RecommendationsModule {
	storage := newRecommendationStorage(db)
	taskPlanner := newTaskPlanner(taskScheduler)

	lastFmRecommendedPlaylists := newLastFmRecommendedPlaylistService(lastFmRecommendations, search, storage, taskPlanner)
	listenBrainzRecommendedPlaylists := newListenBrainzRecommendedPlaylistService(listenBrainzRecommendations, search, storage, taskPlanner)

	return &RecommendationsModule{
		storage:                          storage,
		lastFmRecommendedPlaylists:       lastFmRecommendedPlaylists,
		listenBrainzRecommendedPlaylists: listenBrainzRecommendedPlaylists,
		playlistRefreshCron:              playlistRefreshCron,
	}
}

func (module *RecommendationsModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}

func (module *RecommendationsModule) RegisterSchedules(scheduler *scheduling.TaskScheduler) {
	if module.playlistRefreshCron != nil {
		// todo: session discovery is stupid and should be event-based or something,
		//   but i can't be bothered to waste any more time on this

		sessionDiscoverDelay := 1 * time.Minute
		retryDelay := 15 * time.Minute // todo: config?

		if _, err := scheduler.RegisterTask(TASK_RECOMMENDATIONS_DISCOVER_LASTFM_SESSIONS, nil, time.Now()); err != nil {
			slog.Error(fmt.Sprintf("Failed to register last.fm session discovery task: %s", err.Error()))
		}

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_DISCOVER_LASTFM_SESSIONS,
			sessionDiscoverDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters any) error {
				if err := module.lastFmRecommendedPlaylists.DiscoverSessions(); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, time.Now().Add(sessionDiscoverDelay))
			},
		)

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_DISCOVER_LASTFM_PLAYLISTS,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters DiscoverLastFmRecommendedPlaylistsTask) error {
				if err := module.lastFmRecommendedPlaylists.DiscoverPlaylists(parameters.UserId); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.playlistRefreshCron.Next(time.Now()))
			},
		)

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_SYNC_LASTFM_PLAYLIST,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters SyncLastFmRecommendedPlaylistTask) error {
				if err := module.lastFmRecommendedPlaylists.RefreshPlaylist(parameters.PlaylistId); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.playlistRefreshCron.Next(time.Now()))
			},
		)

		if _, err := scheduler.RegisterTask(TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_SESSIONS, nil, time.Now()); err != nil {
			slog.Error(fmt.Sprintf("Failed to register ListenBrainz session discovery task: %s", err.Error()))
		}

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_SESSIONS,
			sessionDiscoverDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters any) error {
				if err := module.listenBrainzRecommendedPlaylists.DiscoverSessions(); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, time.Now().Add(sessionDiscoverDelay))
			},
		)

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_PLAYLISTS,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters DiscoverListenBrainzRecommendedPlaylistsTask) error {
				if err := module.listenBrainzRecommendedPlaylists.DiscoverPlaylists(parameters.UserId); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.playlistRefreshCron.Next(time.Now()))
			},
		)

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_RECOMMENDATIONS_SYNC_LISTENBRAINZ_PLAYLIST,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters SyncListenBrainzRecommendedPlaylistTask) error {
				if err := module.listenBrainzRecommendedPlaylists.RefreshPlaylist(parameters.PlaylistId); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.playlistRefreshCron.Next(time.Now()))
			},
		)
	}
}
