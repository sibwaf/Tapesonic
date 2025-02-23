package storage

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LastFmSession struct {
	SessionKey string
	Username   string `gorm:"uniqueIndex"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// todo: this will get deleted when tapesonic becomes multi-user
type LastFmSessionStorage struct {
	db *DbHelper
}

func NewLastFmSessionStorage(db *gorm.DB) (*LastFmSessionStorage, error) {
	err := db.AutoMigrate(&LastFmSession{})
	return &LastFmSessionStorage{db: NewDbHelper(db)}, err
}

func (storage *LastFmSessionStorage) Save(session LastFmSession) (LastFmSession, error) {
	return session, storage.db.Clauses(clause.OnConflict{UpdateAll: true}, clause.Returning{}).Create(&session).Error
}

func (storage *LastFmSessionStorage) Find() (*LastFmSession, error) {
	result := LastFmSession{}
	err := storage.db.Order("updated_at DESC").First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result, err
}
