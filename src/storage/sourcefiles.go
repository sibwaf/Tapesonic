package storage

import (
	"errors"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SourceFile struct {
	Id uuid.UUID

	SourceId uuid.UUID `gorm:"uniqueIndex"`
	Source   *Source

	Format string
	Codec  string

	MediaPath string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
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
	sql := `
		INSERT INTO source_files (id, source_id, format, codec, media_path, created_at, updated_at)
		VALUES (@id, @sourceId, @format, @codec, @mediaPath, @createdAt, @updatedAt)
		RETURNING *
	`
	params := map[string]any{
		"id":        file.Id,
		"sourceId":  file.SourceId,
		"format":    file.Format,
		"codec":     file.Codec,
		"mediaPath": file.MediaPath,
		"createdAt": file.CreatedAt,
		"updatedAt": file.UpdatedAt,
	}

	return file, storage.db.Raw(sql, params).First(&file).Error
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
