package storage

import (
	"os"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DataStorage struct {
	db *gorm.DB
}

func NewDataStorage(
	dir string,
) (*DataStorage, error) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(path.Join(dir, "data.sqlite")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(
		&Tape{},
		&TapeTrack{},
	); err != nil {
		return nil, err
	}

	return &DataStorage{
		db: db,
	}, nil
}

func (s *DataStorage) CreateTape(tape *Tape) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(tape).Error; err != nil {
			return err
		}

		if err := tx.Where(
			"tape_id = ? AND tape_track_index > ?",
			tape.Id,
			len(tape.Tracks)-1,
		).Delete(&TapeTrack{}).Error; err != nil {
			return err
		}

		return nil
	})
}

