package remotes

import (
	"tapesonic/artists"
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
	artists *artists.ArtistService,
	librarySyncCron cron.Schedule,
) *RemotesModule {
	remoteStorage := newRemoteStorage(db)
	artworkStorage := newRemoteArtworkStorage(db)
	artistStorage := newRemoteArtistStorage(db)
	albumStorage := newRemoteAlbumStorage(db)
	trackStorage := newRemoteTrackStorage(db)

	planner := newTaskPlanner(taskScheduler)

	service := newRemoteService(remoteStorage, planner)

	subsonicSync := newSubsonicSyncService(
		artists,
		remoteStorage,
		artworkStorage,
		artistStorage,
		albumStorage,
		trackStorage,
	)

	return &RemotesModule{
		RemoteService:   service,
		subsonicSync:    subsonicSync,
		librarySyncCron: librarySyncCron,
	}
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
