package storage

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CachedMuxArtistStorage struct {
	db *DbHelper
}

type CachedMuxArtist struct {
	ServiceName string `gorm:"primaryKey"`
	ArtistId    string `gorm:"primaryKey"`

	Name string

	CachedAt time.Time
}

func NewCachedMuxArtistStorage(db *gorm.DB) (*CachedMuxArtistStorage, error) {
	err := db.AutoMigrate(
		&CachedMuxArtist{},
	)
	return &CachedMuxArtistStorage{db: NewDbHelper(db)}, err
}

func (storage *CachedMuxArtistStorage) Save(item CachedMuxArtist) (*CachedMuxArtist, error) {
	return &item, storage.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error
}

func (storage *CachedMuxArtistStorage) Replace(items []CachedMuxArtist) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&CachedMuxArtist{}).Error; err != nil {
			return err
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}
