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

func (ds *DataStorage) CreateTape(tape *Tape) error {
	return ds.db.Transaction(func(tx *gorm.DB) error {
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

func (ds *DataStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	// todo: get rid of preload
	return result, ds.db.Preload("Tracks").Find(&result).Error
}

func (ds *DataStorage) GetTapeWithoutTracks(id string) (*Tape, error) {
	result := Tape{}
	return &result, ds.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (ds *DataStorage) GetTapeWithTracks(id string) (*Tape, error) {
	result := Tape{}
	return &result, ds.db.Where(&Tape{Id: id}).Preload("Tracks").Take(&result).Error
}

func (ds *DataStorage) GetTapeTrack(tapeId string, index int) (*TapeTrack, error) {
	filter := map[string]any{
		"tape_id":          tapeId,
		"tape_track_index": index, // index can be 0 and gorm ignores default values for fields
	}

	result := TapeTrack{}
	return &result, ds.db.Where(filter).Take(&result).Error
}
