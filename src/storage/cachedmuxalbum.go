package storage

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CachedMuxAlbumStorage struct {
	db *DbHelper
}

type CachedMuxAlbum struct {
	ServiceName string `gorm:"primaryKey"`
	AlbumId     string `gorm:"primaryKey"`

	Artist string
	Title  string

	CachedAt time.Time
}

func NewCachedMuxAlbumStorage(db *gorm.DB) (*CachedMuxAlbumStorage, error) {
	err := db.AutoMigrate(
		&CachedMuxAlbum{},
	)
	return &CachedMuxAlbumStorage{db: NewDbHelper(db)}, err
}

func (storage *CachedMuxAlbumStorage) Save(item CachedMuxAlbum) (*CachedMuxAlbum, error) {
	return &item, storage.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error
}

func (storage *CachedMuxAlbumStorage) Replace(items []CachedMuxAlbum) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&CachedMuxAlbum{}).Error; err != nil {
			return err
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}
