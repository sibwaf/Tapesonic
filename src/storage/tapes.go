package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TAPE_TYPE_ALBUM    = "album"
	TAPE_TYPE_PLAYLIST = "playlist"
)

type TapeStorage struct {
	db *gorm.DB
}

type Tape struct {
	Id uuid.UUID

	Name string
	Type string

	ThumbnailId *uuid.UUID
	Thumbnail   *Thumbnail

	Tracks []TapeToTrack

	Artist     string
	ReleasedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TapeToTrack struct {
	TapeId uuid.UUID `gorm:"primaryKey"`
	Tape   *Tape

	TrackId uuid.UUID `gorm:"primaryKey"`
	Track   *Track

	ListIndex int
}

func NewTapeStorage(db *gorm.DB) (*TapeStorage, error) {
	if err := db.AutoMigrate(&Tape{}, &TapeToTrack{}); err != nil {
		return nil, err
	}

	return &TapeStorage{db: db}, nil
}

func (storage *TapeStorage) Create(tape Tape) (Tape, error) {
	tape.Id = uuid.New()

	for i := range tape.Tracks {
		tape.Tracks[i].ListIndex = i
		tape.Tracks[i].Track = nil
		tape.Tracks[i].Tape = nil
		tape.Tracks[i].TapeId = tape.Id
	}

	err := storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{}).Create(&tape).Error; err != nil {
			return err
		}
		if len(tape.Tracks) > 0 {
			if err := tx.Save(&tape.Tracks).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return tape, err
}

func (storage *TapeStorage) Update(tape Tape) (Tape, error) {
	for i := range tape.Tracks {
		tape.Tracks[i].ListIndex = i
		tape.Tracks[i].Track = nil
		tape.Tracks[i].Tape = nil
		tape.Tracks[i].TapeId = tape.Id
	}

	err := storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{}).Omit("created_at").Save(&tape).Error; err != nil {
			return err
		}
		if err := tx.Where("tape_id = ?", tape.Id).Delete(&TapeToTrack{}).Error; err != nil {
			return err
		}
		if len(tape.Tracks) > 0 {
			if err := tx.Save(&tape.Tracks).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return tape, err
}

func (storage *TapeStorage) DeleteById(id uuid.UUID) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tape_id = ?", id).Delete(&TapeToTrack{}).Error; err != nil {
			return err
		}

		return tx.Delete(&Tape{Id: id}).Error
	})
}

func (storage *TapeStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	return result, storage.db.Order("created_at DESC").Find(&result).Error
}

func (storage *TapeStorage) GetTape(id uuid.UUID) (Tape, error) {
	result := Tape{Id: id}
	return result, storage.db.Find(&result).Error
}
