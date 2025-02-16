package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SourceFile struct {
	Id uuid.UUID

	SourceId uuid.UUID `gorm:"uniqueIndex"`
	Source   *Source

	Format string
	Codec  string

	MediaPath string

	CreatedAt time.Time
	UpdatedAt *time.Time
}

type SourceFileStorage struct {
	db *DbHelper
}

func NewSourceFileStorage(db *gorm.DB) (*SourceFileStorage, error) {
	if err := db.AutoMigrate(&SourceFile{}); err != nil {
		return nil, err
	}

	return &SourceFileStorage{db: NewDbHelper(db)}, nil
}

func (storage *SourceFileStorage) Create(file SourceFile) (SourceFile, error) {
	if file.Id == uuid.Nil {
		file.Id = uuid.New()
	}

	return file, storage.db.Clauses(clause.Returning{}).Create(&file).Error
}

func (storage *SourceFileStorage) DeleteById(id uuid.UUID) error {
	return storage.db.Delete(&SourceFile{Id: id}).Error
}

func (storage *SourceFileStorage) FindBySourceId(sourceId uuid.UUID) (*SourceFile, error) {
	result := SourceFile{}
	if err := storage.db.Where("source_id = ?", sourceId).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &result, nil
}

func (storage *SourceFileStorage) FindBySourceIds(sourceIds []uuid.UUID) ([]SourceFile, error) {
	result := []SourceFile{}
	return result, storage.db.Where("source_id IN ?", sourceIds).Find(&result).Error
}
