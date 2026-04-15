package artworks

import (
	"errors"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtworkStorage struct {
	db *gorm.DB
}

func newArtworkStorage(db *gorm.DB) *ArtworkStorage {
	return &ArtworkStorage{db: db}
}

func (store *ArtworkStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&Artwork{})
}

func (store *ArtworkStorage) Upsert(artwork Artwork) (Artwork, error) {
	return artwork, store.db.Transaction(func(tx *gorm.DB) error {
		query := `
			INSERT INTO artworks (id, deduplication_id, file_path, format, created_at, updated_at)
			VALUES (@id, @deduplicationId, @filePath, @format, @createdAt, @updatedAt)
			ON CONFLICT (deduplication_id) DO UPDATE
			SET file_path = excluded.file_path, format = excluded.format, updated_at = excluded.updated_at
			RETURNING *
		`
		params := map[string]any{
			"id":              artwork.Id,
			"deduplicationId": artwork.DeduplicationId,
			"filePath":        artwork.FilePath,
			"format":          artwork.Format,
			"createdAt":       artwork.CreatedAt,
			"updatedAt":       artwork.UpdatedAt,
		}
		if err := store.db.Raw(query, params).Take(&artwork).Error; err != nil {
			return err
		}

		query = `
			INSERT INTO all_artwork_ids (id, artwork_id)
			VALUES (@id, @id)
			ON CONFLICT (id) DO NOTHING
		`
		params = map[string]any{
			"id": artwork.Id,
		}
		if err := store.db.Exec(query, params).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *ArtworkStorage) GetById(id uuid.UUID) (Artwork, error) {
	sql := `
		SELECT id, deduplication_id, file_path, format, created_at, updated_at
		FROM artworks
		WHERE id = @id
	`
	params := map[string]any{"id": id}

	result := Artwork{}
	err := store.db.Raw(sql, params).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Artwork{}, model.ErrNotFound
	} else {
		return result, err
	}
}
