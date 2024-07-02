package storage

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CachedMuxSongStorage struct {
	db *DbHelper
}

type CachedMuxSong struct {
	ServiceName string `gorm:"primaryKey"`
	SongId      string `gorm:"primaryKey"`

	AlbumId string

	Artist string
	Title  string

	DurationSec int

	CachedAt time.Time
}

func NewCachedMuxSongStorage(db *gorm.DB) (*CachedMuxSongStorage, error) {
	err := db.AutoMigrate(
		&CachedMuxSong{},
	)
	return &CachedMuxSongStorage{db: NewDbHelper(db)}, err
}

func (storage *CachedMuxSongStorage) Save(item CachedMuxSong) (*CachedMuxSong, error) {
	return &item, storage.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error
}

func (storage *CachedMuxSongStorage) Replace(items []CachedMuxSong) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&CachedMuxSong{}).Error; err != nil {
			return err
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}
