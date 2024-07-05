package storage

import (
	"fmt"
	"strings"
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

	SearchArtist string
	SearchTitle  string

	CachedAt time.Time
}

func (album *CachedMuxAlbum) BeforeSave(tx *gorm.DB) (err error) {
	album.SearchArtist = strings.Join(ExtractSearchTerms(album.Artist), " ")
	album.SearchTitle = strings.Join(ExtractSearchTerms(album.Title), " ")
	return nil
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

func (storage *CachedMuxAlbumStorage) Search(query string, count int, offset int) ([]CachedAlbumId, error) {
	result := []CachedAlbumId{}

	filter := MakeTextSearchCondition([]string{"search_artist", "search_title"}, query)
	if filter == "" {
		return result, nil
	}

	sql := fmt.Sprintf(
		`
			SELECT
				service_name AS service_name,
				album_id AS id
			FROM cached_mux_albums
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
