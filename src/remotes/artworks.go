package remotes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteArtwork struct {
	Id uuid.UUID

	RemoteId uuid.UUID
	Remote   Remote

	ArtworkId string
}

type RemoteArtworkStorage struct {
	db *gorm.DB
}

func newRemoteArtworkStorage(db *gorm.DB) *RemoteArtworkStorage {
	return &RemoteArtworkStorage{db: db}
}

func (store *RemoteArtworkStorage) Upsert(artwork RemoteArtwork) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		sql := `
			INSERT INTO remote_artworks (id, remote_id, artwork_id)
			VALUES (@id, @remoteId, @artworkId)
			ON CONFLICT (remote_id, artwork_id) DO UPDATE
			SET id = remote_artworks.id
			RETURNING id, remote_id, artwork_id
		`
		params := map[string]any{
			"id":        artwork.Id,
			"remoteId":  artwork.RemoteId,
			"artworkId": artwork.ArtworkId,
		}

		if err := store.db.Raw(sql, params).Take(&artwork).Error; err != nil {
			return err
		}

		sql = `
			INSERT INTO all_artwork_ids (id, remote_artwork_id)
			VALUES (@id, @id)
			ON CONFLICT (id) DO NOTHING
		`
		params = map[string]any{
			"id": artwork.Id,
		}
		if err := store.db.Exec(sql, params).Error; err != nil {
			return err
		}

		return nil
	})
}
