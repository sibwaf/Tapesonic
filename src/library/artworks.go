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

func (store *ArtworkStorage) PrepareDatabase() error {
	sql := `
		CREATE TABLE IF NOT EXISTS all_artwork_ids (
			id TEXT NOT NULL PRIMARY KEY,
			artwork_id TEXT,
			remote_artwork_id TEXT,

			CONSTRAINT all_artwork_ids_artwork_id_fk FOREIGN KEY (artwork_id) REFERENCES artworks (id),
			CONSTRAINT all_artwork_ids_remote_artwork_id_fk FOREIGN KEY (remote_artwork_id) REFERENCES remote_artworks (id)
		)
	`
	if err := store.db.Exec(sql).Error; err != nil {
		return err
	}

	sql = `
		CREATE VIEW all_artworks (
			id,
			remote_id,
			remote_artwork_id
		) AS

		SELECT
			artworks.id AS id,
			NULL AS remote_id,
			NULL AS remote_artwork_id
		FROM artworks

		UNION ALL

		SELECT
			remote_artworks.id AS id,
			remote_artworks.remote_id AS remote_id,
			remote_artworks.artwork_id AS remote_artwork_id
		FROM remote_artworks
	`

	return store.db.Exec(sql).Error
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
