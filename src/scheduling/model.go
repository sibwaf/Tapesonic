package scheduling

import (
	"tapesonic/util"

	"github.com/google/uuid"
)

type ScheduledTask struct {
	Id uuid.UUID

	Type       string
	Parameters string

	RunAt          util.TimestampWrapper
	RunAttemptedAt *util.TimestampWrapper
	RunSucceededAt *util.TimestampWrapper
}

type TaskRunner[T any] func(scheduler *TaskScheduler, task ScheduledTask, parameters T) error
