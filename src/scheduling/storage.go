package scheduling

import (
	"errors"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStorage struct {
	db *gorm.DB
}

func newTaskStorage(db *gorm.DB) *TaskStorage {
	return &TaskStorage{db: db}
}

func (store *TaskStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&ScheduledTask{})
}

func (store *TaskStorage) FindTaskByParameters(taskType string, parameters string) (*ScheduledTask, error) {
	sql := `
		SELECT id, type, parameters, run_at, run_attempted_at, run_succeeded_at
		FROM scheduled_tasks
		WHERE type = @type AND parameters = @parameters
		LIMIT 1
	`
	params := map[string]any{
		"type":       taskType,
		"parameters": parameters,
	}

	result := ScheduledTask{}
	err := store.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}

func (store *TaskStorage) UpsertTaskWithoutReschedule(task ScheduledTask) (ScheduledTask, error) {
	sql := `
		INSERT INTO scheduled_tasks (id, type, parameters, run_at)
		VALUES (@id, @type, @parameters, @runAt)
		ON CONFLICT (type, parameters) DO UPDATE
		SET run_at = scheduled_tasks.run_at
		RETURNING *
	`
	params := map[string]any{
		"id":         task.Id,
		"type":       task.Type,
		"parameters": task.Parameters,
		"runAt":      task.RunAt,
	}
	return task, store.db.Raw(sql, params).First(&task).Error
}

func (store *TaskStorage) FindTaskToRun(taskType string, maxRunAt time.Time, maxRunAttemptedAt time.Time) (*ScheduledTask, error) {
	sql := `
		SELECT *
		FROM scheduled_tasks
		WHERE type = @type
			AND run_at < @maxRunAt
			AND (run_succeeded_at IS NULL OR run_succeeded_at < run_at)
			AND (run_attempted_at IS NULL OR run_attempted_at < @maxRunAttemptedAt OR run_attempted_at < run_at)
		ORDER BY run_attempted_at ASC NULLS FIRST, run_at ASC
		LIMIT 1
	`
	params := map[string]any{
		"type":              taskType,
		"maxRunAt":          util.NewTimestampWrapper(maxRunAt),
		"maxRunAttemptedAt": util.NewTimestampWrapper(maxRunAttemptedAt),
	}

	result := ScheduledTask{}
	err := store.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}

func (store *TaskStorage) RescheduleTask(id uuid.UUID, runAt time.Time) error {
	sql := `
		UPDATE scheduled_tasks
		SET run_at = @runAt
		WHERE id = @id
	`
	params := map[string]any{
		"id":    id,
		"runAt": util.NewTimestampWrapper(runAt),
	}
	return store.db.Exec(sql, params).Error
}

func (store *TaskStorage) UpdateTaskTimestamps(id uuid.UUID, runAttemptedAt time.Time, runSucceededAt *time.Time) error {
	sql := `
		UPDATE scheduled_tasks
		SET run_attempted_at = @runAttemptedAt, run_succeeded_at = @runSucceededAt
		WHERE id = @id
	`
	params := map[string]any{
		"id":             id,
		"runAttemptedAt": util.NewTimestampWrapper(runAttemptedAt),
		"runSucceededAt": util.NewTimestampWrapperOrNull(runSucceededAt),
	}
	return store.db.Exec(sql, params).Error
}

func (store *TaskStorage) DeleteTask(id uuid.UUID) error {
	sql := `DELETE FROM scheduled_tasks WHERE id = @id`
	params := map[string]any{"id": id}
	return store.db.Exec(sql, params).Error
}
