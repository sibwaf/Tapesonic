package storage

import (
	"fmt"
	"strings"
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
	Album  string
	Title  string

	DurationSec int

	SearchArtist string
	SearchAlbum  string
	SearchTitle  string

	CachedAt time.Time
}

func (song *CachedMuxSong) BeforeSave(tx *gorm.DB) (err error) {
	song.SearchArtist = strings.Join(ExtractSearchTerms(song.Artist), " ")
	song.SearchAlbum = strings.Join(ExtractSearchTerms(song.Album), " ")
	song.SearchTitle = strings.Join(ExtractSearchTerms(song.Title), " ")
	return nil
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

func (storage *CachedMuxSongStorage) GetById(serviceName string, songId string) (*CachedMuxSong, error) {
	result := CachedMuxSong{
		ServiceName: serviceName,
		SongId:      songId,
	}
	return &result, storage.db.First(&result).Error
}

func (storage *CachedMuxSongStorage) Search(query string, count int, offset int) ([]CachedSongId, error) {
	result := []CachedSongId{}

	filter := MakeTextSearchCondition([]string{"search_artist", "search_album", "search_title"}, query)
	if filter == "" {
		return result, nil
	}

	sql := fmt.Sprintf(
		`
			SELECT
				service_name AS service_name,
				song_id AS id
			FROM cached_mux_songs
			WHERE %s
			ORDER BY id
			LIMIT %d OFFSET %d
		`,
		filter,
		count,
		offset,
	)

	return result, storage.db.Raw(sql).Find(&result).Error
}
