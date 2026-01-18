package storage

import (
	"errors"
	"tapesonic/util"
	"time"

	"gorm.io/gorm"
)

type YtdlpMetadataCacheItem struct {
	Url string `gorm:"uniqueIndex"`

	Metadata string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type YtdlpMetadataStorage struct {
	db *DbHelper
}

func NewYtdlpMetadataStorage(db *gorm.DB) (*YtdlpMetadataStorage, error) {
	if err := db.AutoMigrate(&YtdlpMetadataCacheItem{}); err != nil {
		return nil, err
	}

	return &YtdlpMetadataStorage{db: NewDbHelper(db)}, nil
}

func (s *YtdlpMetadataStorage) Upsert(url string, metadata string) error {
	query := `
		INSERT INTO ytdlp_metadata_cache_items (url, metadata, created_at, updated_at)
		VALUES (@url, @metadata, @createdAt, @updatedAt)
		ON CONFLICT (url) DO UPDATE
		SET metadata = excluded.metadata, updated_at = excluded.updated_at
	`
	params := map[string]any{
		"url":       url,
		"metadata":  metadata,
		"createdAt": util.NewTimestampWrapper(time.Now()),
		"updatedAt": util.NewTimestampWrapper(time.Now()),
	}

	return s.db.Exec(query, params).Error
}

func (s *YtdlpMetadataStorage) Find(url string, minUpdatedAt time.Time) (*YtdlpMetadataCacheItem, error) {
	sql := `
		SELECT *
		FROM ytdlp_metadata_cache_items
		WHERE url = @url AND updated_at >= @minUpdatedAt
	`
	params := map[string]any{
		"url":          url,
		"minUpdatedAt": util.NewTimestampWrapper(minUpdatedAt),
	}

	result := YtdlpMetadataCacheItem{}
	err := s.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}
