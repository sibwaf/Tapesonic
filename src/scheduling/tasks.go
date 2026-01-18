package scheduling

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type TaskScheduler struct {
	tasks         *TaskStorage
	looper        *Looper
	checkInterval time.Duration
}

func newTaskScheduler(tasks *TaskStorage, looper *Looper, checkInterval time.Duration) *TaskScheduler {
	return &TaskScheduler{tasks: tasks, looper: looper, checkInterval: checkInterval}
}

func (svc *TaskScheduler) FindTask(taskType string, parameters any) (*ScheduledTask, error) {
	parametersJson, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}

	return svc.tasks.FindTaskByParameters(taskType, string(parametersJson))
}

func (svc *TaskScheduler) RegisterTask(taskType string, parameters any, runAt time.Time) (ScheduledTask, error) {
	parametersJson, err := json.Marshal(parameters)
	if err != nil {
		return ScheduledTask{}, err
	}

	task := ScheduledTask{
		Id:         uuid.New(),
		Type:       taskType,
		Parameters: string(parametersJson),
		RunAt:      util.NewTimestampWrapper(runAt),
	}

	return svc.tasks.UpsertTaskWithoutReschedule(task)
}

func (svc *TaskScheduler) RescheduleTask(id uuid.UUID, runAt time.Time) error {
	return svc.tasks.RescheduleTask(id, runAt)
}

func (svc *TaskScheduler) UnregisterTask(id uuid.UUID) error {
	return svc.tasks.DeleteTask(id)
}

func SubscribeTaskRunner[T any](scheduler *TaskScheduler, taskType string, retryDelay time.Duration, runner TaskRunner[T]) {
	// todo: no-sleep spin while there are more tasks ready to run
	// todo: multiple concurrent task runs

	scheduler.looper.RegisterInterval(taskType, scheduler.checkInterval, func() error {
		now := time.Now()
		task, err := scheduler.tasks.FindTaskToRun(taskType, now, now.Add(-retryDelay))
		if err != nil {
			return err
		}
		if task == nil {
			return nil
		}

		task.RunAttemptedAt = util.NewTimestampWrapperOrNull(&now)

		defer func() {
			err := scheduler.tasks.UpdateTaskTimestamps(
				task.Id,
				task.RunAttemptedAt.Unwrap(),
				task.RunSucceededAt.UnwrapNullable(),
			)
			if err != nil {
				slog.Error(fmt.Sprintf("Failed to update timestamps after scheduled task id=%s execution: %s", task.Id, err.Error()))
			}
		}()

		var parameters T
		if err := json.Unmarshal([]byte(task.Parameters), &parameters); err != nil {
			return err
		}

		runErr := runner(scheduler, *task, parameters)

		if runErr == nil {
			task.RunSucceededAt = util.NewTimestampWrapperOrNull(&now)
		}

		return runErr
	})
}
