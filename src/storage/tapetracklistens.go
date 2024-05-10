package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TapeTrackListensStorage struct {
	db *DbHelper
}

type TapeTrackListens struct {
	Id int64

	TapeTrackId uuid.UUID `gorm:"uniqueIndex"`
	TapeTrack   *TapeTrack

	ListenCount int

	LastListenedAt time.Time
}

func NewTapeTrackListensStorage(db *gorm.DB) (*TapeTrackListensStorage, error) {
	err := db.AutoMigrate(
		&TapeTrackListens{},
	)
	return &TapeTrackListensStorage{db: NewDbHelper(db)}, err
}

func (storage *TapeTrackListensStorage) Record(tapeTrackId uuid.UUID, listenedAt time.Time, incrementListenCount bool) error {
	return storage.db.ExclusiveTransaction(func(tx *gorm.DB) error {
		item := TapeTrackListens{}
		if err := tx.Where(&TapeTrackListens{TapeTrackId: tapeTrackId}).Find(&item).Error; err != nil {
			return err
		}

		item.TapeTrackId = tapeTrackId
		if incrementListenCount {
			item.ListenCount += 1
		}
		if listenedAt.After(item.LastListenedAt) {
			item.LastListenedAt = listenedAt
		}

		return tx.Save(&item).Error
	})
}
