package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Thumbnail struct {
	Id uuid.UUID

	DeduplicationId string `gorm:"uniqueIndex"`

	FilePath string
	Format   string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ThumbnailStorage struct {
	db *DbHelper
}

func NewThumbnailStorage(db *gorm.DB) (*ThumbnailStorage, error) {
	return &ThumbnailStorage{db: NewDbHelper(db)}, db.AutoMigrate(&Thumbnail{})
}

func (s *ThumbnailStorage) Upsert(thumbnail Thumbnail) (Thumbnail, error) {
	if thumbnail.Id == uuid.Nil {
		thumbnail.Id = uuid.New()
	}

	return thumbnail, s.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "deduplication_id"}},
			UpdateAll: true,
		},
		clause.Returning{},
	).Create(&thumbnail).Error
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
