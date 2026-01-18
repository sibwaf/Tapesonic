package scheduling

import (
	"time"

	"gorm.io/gorm"
)

type SchedulingModule struct {
	storage       *TaskStorage
	Looper        *Looper
	TaskScheduler *TaskScheduler
}

func NewSchedulingModule(db *gorm.DB, taskCheckInterval time.Duration) *SchedulingModule {
	storage := newTaskStorage(db)
	looper := newLooper()
	taskScheduler := newTaskScheduler(storage, looper, taskCheckInterval)

	return &SchedulingModule{
		storage:       storage,
		Looper:        looper,
		TaskScheduler: taskScheduler,
	}
}

func (module *SchedulingModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
