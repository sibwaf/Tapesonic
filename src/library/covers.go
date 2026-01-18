package library

import (
	"errors"
	"tapesonic/model"

	"gorm.io/gorm"
)

type CoverStorage struct {
	db *gorm.DB
}

func newCoverStorage(db *gorm.DB) *CoverStorage {
	return &CoverStorage{db: db}
}

func (store *CoverStorage) PrepareDatabase() error {
	sql := `
		CREATE VIEW all_covers (
			id,
			remote_id,
			remote_cover_id
		) AS

		SELECT
			thumbnails.id AS id,
			NULL AS remote_id,
			NULL AS remote_cover_id
		FROM thumbnails

		UNION ALL

		SELECT
			remote_covers.id AS id,
			remote_covers.remote_id AS remote_id,
			remote_covers.cover_id AS remote_cover_id
		FROM remote_covers
	`

	return store.db.Exec(sql).Error
}

func (store *CoverStorage) FindCoverById(id string) (*model.LibraryCover, error) {
	result := model.LibraryCover{}

	params := map[string]any{
		"id": id,
	}

	err := store.db.Raw("SELECT * FROM all_covers WHERE id = @id LIMIT 1", params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}
