package storage

import (
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thumbnail struct {
	Id uuid.UUID

	DeduplicationId string `gorm:"uniqueIndex"`

	FilePath string
	Format   string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type ThumbnailStorage struct {
	db *DbHelper
}

func NewThumbnailStorage(db *gorm.DB) (*ThumbnailStorage, error) {
	return &ThumbnailStorage{db: NewDbHelper(db)}, db.AutoMigrate(&Thumbnail{})
}

func (s *ThumbnailStorage) Upsert(thumbnail Thumbnail) (Thumbnail, error) {
	query := `
		INSERT INTO thumbnails (id, deduplication_id, file_path, format, created_at, updated_at)
		VALUES (@id, @deduplicationId, @filePath, @format, @createdAt, @updatedAt)
		ON CONFLICT (deduplication_id) DO UPDATE
		SET file_path = excluded.file_path, format = excluded.format, updated_at = excluded.updated_at
		RETURNING *
	`
	params := map[string]any{
		"id":              thumbnail.Id,
		"deduplicationId": thumbnail.DeduplicationId,
		"filePath":        thumbnail.FilePath,
		"format":          thumbnail.Format,
		"createdAt":       thumbnail.CreatedAt,
		"updatedAt":       thumbnail.UpdatedAt,
	}

	return thumbnail, s.db.Raw(query, params).First(&thumbnail).Error
}

func (s *ThumbnailStorage) Search(sourceIds []uuid.UUID) ([]Thumbnail, error) {
	query := `
		SELECT *
		FROM (
			SELECT
				thumbnails.*,
				row_number() OVER (PARTITION BY thumbnails.id) AS rownum
			FROM thumbnails
			JOIN sources ON sources.thumbnail_id = thumbnails.id
			WHERE sources.id IN @sourceIds
		)
		WHERE rownum = 1
	`

	args := map[string]any{
		"sourceIds": sourceIds,
	}

	result := []Thumbnail{}
	return result, s.db.Raw(query, args).Find(&result).Error
}

func (s *ThumbnailStorage) GetById(id uuid.UUID) (Thumbnail, error) {
	result := Thumbnail{Id: id}
	return result, s.db.Find(&result).Error
}
