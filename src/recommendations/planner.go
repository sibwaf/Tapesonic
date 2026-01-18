package recommendations

import (
	"tapesonic/scheduling"
	"time"

	"github.com/google/uuid"
)

type TaskPlanner struct {
	scheduler *scheduling.TaskScheduler
}

func newTaskPlanner(
	scheduler *scheduling.TaskScheduler,
) *TaskPlanner {
	return &TaskPlanner{scheduler: scheduler}
}

// todo: unregister unneeded tasks by events (playlist deleted, session logged out, ...)?

func (planner *TaskPlanner) ScheduleLastFmPlaylistDiscovery(userId uuid.UUID) error {
	_, err := planner.scheduler.RegisterTask(
		TASK_RECOMMENDATIONS_DISCOVER_LASTFM_PLAYLISTS,
		DiscoverLastFmRecommendedPlaylistsTask{UserId: userId},
		time.Now(),
	)
	return err
}

func (planner *TaskPlanner) CancelLastFmPlaylistDiscovery(userId uuid.UUID) error {
	task, err := planner.scheduler.FindTask(
		TASK_RECOMMENDATIONS_DISCOVER_LASTFM_PLAYLISTS,
		DiscoverLastFmRecommendedPlaylistsTask{UserId: userId},
	)
	if task == nil || err != nil {
		return err
	}
	return planner.scheduler.UnregisterTask(task.Id)
}

func (planner *TaskPlanner) ScheduleLastFmPlaylistSync(playlistId uuid.UUID) error {
	_, err := planner.scheduler.RegisterTask(
		TASK_RECOMMENDATIONS_SYNC_LASTFM_PLAYLIST,
		SyncLastFmRecommendedPlaylistTask{PlaylistId: playlistId},
		time.Now(),
	)
	return err
}

func (planner *TaskPlanner) CancelLastFmPlaylistSync(playlistId uuid.UUID) error {
	task, err := planner.scheduler.FindTask(
		TASK_RECOMMENDATIONS_SYNC_LASTFM_PLAYLIST,
		SyncLastFmRecommendedPlaylistTask{PlaylistId: playlistId},
	)
	if task == nil || err != nil {
		return err
	}
	return planner.scheduler.UnregisterTask(task.Id)
}

func (planner *TaskPlanner) ScheduleListenBrainzPlaylistDiscovery(userId uuid.UUID) error {
	_, err := planner.scheduler.RegisterTask(
		TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_PLAYLISTS,
		DiscoverListenBrainzRecommendedPlaylistsTask{UserId: userId},
		time.Now(),
	)
	return err
}

func (planner *TaskPlanner) CancelListenBrainzPlaylistDiscovery(userId uuid.UUID) error {
	task, err := planner.scheduler.FindTask(
		TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_PLAYLISTS,
		DiscoverListenBrainzRecommendedPlaylistsTask{UserId: userId},
	)
	if task == nil || err != nil {
		return err
	}
	return planner.scheduler.UnregisterTask(task.Id)
}

func (planner *TaskPlanner) ScheduleListenBrainzPlaylistSync(playlistId uuid.UUID) error {
	_, err := planner.scheduler.RegisterTask(
		TASK_RECOMMENDATIONS_SYNC_LISTENBRAINZ_PLAYLIST,
		SyncListenBrainzRecommendedPlaylistTask{PlaylistId: playlistId},
		time.Now(),
	)
	return err
}

func (planner *TaskPlanner) CancelListenBrainzPlaylistSync(playlistId uuid.UUID) error {
	task, err := planner.scheduler.FindTask(
		TASK_RECOMMENDATIONS_SYNC_LISTENBRAINZ_PLAYLIST,
		SyncListenBrainzRecommendedPlaylistTask{PlaylistId: playlistId},
	)
	if task == nil || err != nil {
		return err
	}
	return planner.scheduler.UnregisterTask(task.Id)
}
