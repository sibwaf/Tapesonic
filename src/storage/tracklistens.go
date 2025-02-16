package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackListensStorage struct {
	db *DbHelper
}

type TrackListens struct {
	Id int64

	TrackId uuid.UUID `gorm:"uniqueIndex"`
	Track   *Track

	ListenCount int

	LastListenedAt time.Time
}

func NewTrackListensStorage(db *gorm.DB) (*TrackListensStorage, error) {
	err := db.AutoMigrate(
		&TrackListens{},
	)
	return &TrackListensStorage{db: NewDbHelper(db)}, err
}

func (storage *TrackListensStorage) Record(trackId uuid.UUID, listenedAt time.Time, incrementListenCount bool) error {
	return storage.db.ExclusiveTransaction(func(tx *gorm.DB) error {
		item := TrackListens{}
		if err := tx.Where(&TrackListens{TrackId: trackId}).Find(&item).Error; err != nil {
			return err
		}

		item.TrackId = trackId
		if incrementListenCount {
			item.ListenCount += 1
		}
		if listenedAt.After(item.LastListenedAt) {
			item.LastListenedAt = listenedAt
		}

		return tx.Save(&item).Error
	})
}
