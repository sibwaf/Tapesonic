package storage

import (
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

	DurationSec int
}

func NewCachedMuxSongStorage(db *gorm.DB) (*CachedMuxSongStorage, error) {
	err := db.AutoMigrate(
		&CachedMuxSong{},
	)
	return &CachedMuxSongStorage{db: NewDbHelper(db)}, err
}

func (storage *CachedMuxSongStorage) Save(serviceName string, songId string, albumId string, durationSec int) (*CachedMuxSong, error) {
	song := CachedMuxSong{
		ServiceName: serviceName,
		SongId:      songId,
		AlbumId:     albumId,
		DurationSec: durationSec,
	}
	return &song, storage.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&song).Error
}
