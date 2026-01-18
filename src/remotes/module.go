package remotes

import (
	"tapesonic/scheduling"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type RemotesModule struct {
	RemoteService   *RemoteService
	subsonicSync    *SubsonicSyncService
	librarySyncCron cron.Schedule
}

func NewRemotesModule(
	db *gorm.DB,
	taskScheduler *scheduling.TaskScheduler,
	librarySyncCron cron.Schedule,
) (*RemotesModule, error) {
	remoteStorage, err := newRemoteStorage(db)
	if err != nil {
		return nil, err
	}

	coverStorage, err := newRemoteCoverStorage(db)
	if err != nil {
		return nil, err
	}

	artistStorage, err := newRemoteArtistStorage(db)
	if err != nil {
		return nil, err
	}

	albumStorage, err := newRemoteAlbumStorage(db)
	if err != nil {
		return nil, err
	}

	trackStorage, err := newRemoteTrackStorage(db)
	if err != nil {
		return nil, err
	}

	planner := newTaskPlanner(taskScheduler)

	service := newRemoteService(remoteStorage, planner)

	subsonicSync := newSubsonicSyncService(
		remoteStorage,
		coverStorage,
		artistStorage,
		albumStorage,
		trackStorage,
	)

	return &RemotesModule{
		RemoteService:   service,
		subsonicSync:    subsonicSync,
		librarySyncCron: librarySyncCron,
	}, nil
}

func (module *RemotesModule) RegisterSchedules(scheduler *scheduling.TaskScheduler) {
	if module.librarySyncCron != nil {
		retryDelay := 15 * time.Minute // todo: config?

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_REMOTES_SYNC_SUBSONIC_LIBRARY,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters SyncSubsonicLibraryTask) error {
				if err := module.subsonicSync.SyncLibrary(parameters.UserId, parameters.RemoteId); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.librarySyncCron.Next(time.Now()))
			},
		)
	}
}
