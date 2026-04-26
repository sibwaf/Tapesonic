package sources

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileStorage struct {
	db *gorm.DB
}

func newFileStorage(db *gorm.DB) *FileStorage {
	return &FileStorage{db: db}
}

func (store *FileStorage) Create(file SourceFile) (SourceFile, error) {
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

	return file, store.db.Raw(sql, params).First(&file).Error
}

func (store *FileStorage) DeleteById(id uuid.UUID) error {
	return store.db.Delete(&SourceFile{Id: id}).Error
}

func (store *FileStorage) FindBySourceId(sourceId uuid.UUID) (*SourceFile, error) {
	result := SourceFile{}
	err := store.db.Where("source_id = ?", sourceId).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}

func (store *FileStorage) FindBySourceIds(sourceIds []uuid.UUID) ([]SourceFile, error) {
	result := []SourceFile{}
	return result, store.db.Where("source_id IN ?", sourceIds).Find(&result).Error
}
