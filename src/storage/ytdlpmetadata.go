package storage

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type YtdlpMetadataCacheItem struct {
	Url string `gorm:"uniqueIndex"`

	Metadata string

	CreatedAt time.Time
	UpdatedAt time.Time
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
	item := YtdlpMetadataCacheItem{Url: url, Metadata: metadata}
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error
}

func (s *YtdlpMetadataStorage) Find(url string, minUpdatedAt time.Time) (*YtdlpMetadataCacheItem, error) {
	result := YtdlpMetadataCacheItem{}
	err := s.db.Where("url = ? AND updated_at >= ?", url, minUpdatedAt).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}
