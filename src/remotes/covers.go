package remotes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteCover struct {
	Id uuid.UUID

	RemoteId uuid.UUID `gorm:"uniqueIndex:remote_cover_uniq"`
	Remote   Remote

	CoverId string `gorm:"uniqueIndex:remote_cover_uniq"`
}

type RemoteCoverStorage struct {
	db *gorm.DB
}

func newRemoteCoverStorage(db *gorm.DB) (*RemoteCoverStorage, error) {
	if err := db.AutoMigrate(&RemoteCover{}); err != nil {
		return nil, err
	}

	return &RemoteCoverStorage{db: db}, nil
}

func (store *RemoteCoverStorage) Upsert(remoteCover RemoteCover) error {
	sql := `
		INSERT INTO remote_covers (id, remote_id, cover_id)
		VALUES (@id, @remoteId, @coverId)
		ON CONFLICT (remote_id, cover_id) DO NOTHING
	`
	params := map[string]any{
		"id":       remoteCover.Id,
		"remoteId": remoteCover.RemoteId,
		"coverId":  remoteCover.CoverId,
	}

	return store.db.Exec(sql, params).Error
}
