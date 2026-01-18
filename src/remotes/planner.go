package remotes

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

func (planner *TaskPlanner) ScheduleSubsonicLibrarySync(userId uuid.UUID, remoteId uuid.UUID) error {
	_, err := planner.scheduler.RegisterTask(
		TASK_REMOTES_SYNC_SUBSONIC_LIBRARY,
		SyncSubsonicLibraryTask{UserId: userId, RemoteId: remoteId},
		time.Now(),
	)
	return err
}

func (planner *TaskPlanner) CancelSubsonicLibrarySync(userId uuid.UUID, remoteId uuid.UUID) error {
	task, err := planner.scheduler.FindTask(
		TASK_REMOTES_SYNC_SUBSONIC_LIBRARY,
		SyncSubsonicLibraryTask{UserId: userId, RemoteId: remoteId},
	)
	if err != nil {
		return err
	}

	if task != nil {
		return planner.scheduler.UnregisterTask(task.Id)
	} else {
		return nil
	}
}
