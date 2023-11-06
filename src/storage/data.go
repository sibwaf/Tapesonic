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

	return &DataStorage{
		db: db,
	}, nil
}

