package storage

import (
	"os"
	"path"

	"github.com/google/uuid"
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

func (ds *DataStorage) UpsertTape(tape *Tape) error {
	for index, track := range tape.Tracks {
		track.TrackIndex = index
	}

	return ds.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tracks").Save(tape).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(tape).Association("Tracks").Unscoped().Replace(tape.Tracks); err != nil {
			return err
		}

		return nil
	})
}

func (ds *DataStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	// todo: get rid of preload
	return result, ds.db.Preload("Tracks").Find(&result).Error
}

func (ds *DataStorage) GetTapeWithoutTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}
	return &result, ds.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (ds *DataStorage) GetTapeWithTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}
	return &result, ds.db.Where(&Tape{Id: id}).Preload("Tracks").Take(&result).Error
}

func (ds *DataStorage) GetTapeTrack(id uuid.UUID) (*TapeTrack, error) {
	result := TapeTrack{}
	return &result, ds.db.Where(&TapeTrack{Id: id}).Take(&result).Error
}
