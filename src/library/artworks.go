package library

import (
	"errors"
	"tapesonic/model"

	"gorm.io/gorm"
)

type ArtworkStorage struct {
	db *gorm.DB
}

func newArtworkStorage(db *gorm.DB) *ArtworkStorage {
	return &ArtworkStorage{db: db}
}

func (store *ArtworkStorage) FindById(id string) (*model.LibraryArtwork, error) {
	result := model.LibraryArtwork{}

	params := map[string]any{
		"id": id,
	}

	err := store.db.Raw("SELECT * FROM all_artworks WHERE id = @id LIMIT 1", params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}
