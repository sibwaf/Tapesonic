package storage

import (
	"fmt"
	"strings"
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

	SearchName string

	CachedAt time.Time
}

func (artist *CachedMuxArtist) BeforeSave(tx *gorm.DB) (err error) {
	artist.SearchName = strings.Join(ExtractSearchTerms(artist.Name), " ")
	return nil
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

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (storage *CachedMuxArtistStorage) Search(query string, count int, offset int) ([]CachedArtistId, error) {
	result := []CachedArtistId{}

	filter := MakeTextSearchCondition([]string{"search_name"}, query)
	if filter == "" {
		return result, nil
	}

	sql := fmt.Sprintf(
		`
			SELECT
				service_name AS service_name,
				artist_id AS id
			FROM cached_mux_artists
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
