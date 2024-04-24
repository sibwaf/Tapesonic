package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	STATUS_PENDING   = "PENDING"
	STATUS_COMPLETED = "COMPLETED"
)

type ImportQueueStorage struct {
	db *gorm.DB
}

type ImportQueueItem struct {
	Id uuid.UUID

	Url string

	TapeId *uuid.UUID
	Tape   *Tape

	Status string

	CreatedAt time.Time
	UpdatedAt time.Time
	FailedAt  *time.Time
}

func NewImportQueueStorage(db *gorm.DB) (*ImportQueueStorage, error) {
	err := db.AutoMigrate(
		&ImportQueueItem{},
	)
	return &ImportQueueStorage{db: db}, err
}

func (storage *ImportQueueStorage) Enqueue(url string) (*ImportQueueItem, error) {
	item := ImportQueueItem{
		Id:     uuid.New(),
		Url:    url,
		Status: STATUS_PENDING,
	}
	err := storage.db.Create(&item).Error
	return &item, err
}

func (storage *ImportQueueStorage) Delete(itemId uuid.UUID) error {
	return storage.db.Delete(&ImportQueueItem{Id: itemId}).Error
}

func (storage *ImportQueueStorage) Fail(itemId uuid.UUID) error {
	return storage.db.Model(&ImportQueueItem{Id: itemId}).Update("failed_at", time.Now()).Error
}

func (storage *ImportQueueStorage) Complete(itemId uuid.UUID, tapeId uuid.UUID) error {
	return storage.db.Model(&ImportQueueItem{Id: itemId}).Update("tape_id", tapeId).Update("status", STATUS_COMPLETED).Update("failed_at", nil).Error
}

func (storage *ImportQueueStorage) GetAllEnqueued() ([]ImportQueueItem, error) {
	result := []ImportQueueItem{}
	return result, storage.db.Where("status != ?", STATUS_COMPLETED).Order("created_at ASC").Find(&result).Error
}

func (storage *ImportQueueStorage) FetchNext(cooldown time.Duration) (*ImportQueueItem, error) {
	result := []ImportQueueItem{}
	err := storage.db.Where("status != ? AND (failed_at IS NULL OR failed_at < ?)", STATUS_COMPLETED, time.Now().Add(-cooldown)).Order("updated_at ASC").Limit(1).Find(&result).Error
	if err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return &result[0], nil
	} else {
		return nil, nil
	}
}
