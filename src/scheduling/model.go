package scheduling

import (
	"tapesonic/util"

	"github.com/google/uuid"
)

type ScheduledTask struct {
	Id uuid.UUID

	Type       string `gorm:"uniqueIndex:scheduled_task_uniq"`
	Parameters string `gorm:"uniqueIndex:scheduled_task_uniq"`

	RunAt          util.TimestampWrapper
	RunAttemptedAt *util.TimestampWrapper
	RunSucceededAt *util.TimestampWrapper
}

type TaskRunner[T any] func(scheduler *TaskScheduler, task ScheduledTask, parameters T) error
