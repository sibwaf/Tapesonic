package remotes

import (
	"encoding/json"
	"errors"
	"tapesonic/model"
	"tapesonic/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteStorage struct {
	db *storage.DbHelper
}

func newRemoteStorage(db *gorm.DB) (*RemoteStorage, error) {
	return &RemoteStorage{db: storage.NewDbHelper(db)}, db.AutoMigrate(&Remote{}, &RemoteToUser{})
}

func (storage *RemoteStorage) GetAll() ([]Remote, error) {
	result := []Remote{}
	return result, storage.db.Find(&result).Error
}

func (storage *RemoteStorage) FindById(id uuid.UUID) (*Remote, error) {
	sql := `
		SELECT *
		FROM remotes
		WHERE id = @id
	`
	params := map[string]any{"id": id}

	result := Remote{}

	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result, err
}

func (storage *RemoteStorage) Create(remote Remote) (Remote, error) {
	sql := `
		INSERT INTO remotes (id, name, type, url, is_scrobble_replication_enabled, is_external_scrobbling_enabled, created_at)
		VALUES (@id, @name, @type, @url, @isScrobbleReplicationEnabled, @isExternalScrobblingEnabled, @createdAt)
		RETURNING *
	`
	params := map[string]any{
		"id":                           remote.Id,
		"name":                         remote.Name,
		"type":                         remote.Type,
		"url":                          remote.Url,
		"isScrobbleReplicationEnabled": remote.IsScrobbleReplicationEnabled,
		"isExternalScrobblingEnabled":  remote.IsExternalScrobblingEnabled,
		"createdAt":                    remote.CreatedAt,
	}

	return remote, storage.db.Raw(sql, params).First(&remote).Error
}

func (storage *RemoteStorage) UpdateSettings(remoteId uuid.UUID, settings RemoteSettings) (Remote, error) {
	sql := `
		UPDATE remotes
		SET
			name = @name,
			type = @type,
			url = @url,
			is_scrobble_replication_enabled = @isScrobbleReplicationEnabled,
			is_external_scrobbling_enabled = @isExternalScrobblingEnabled
		WHERE id = @id
		RETURNING *
	`
	params := map[string]any{
		"id":                           remoteId,
		"name":                         settings.Name,
		"type":                         settings.Type,
		"url":                          settings.Url,
		"isScrobbleReplicationEnabled": settings.IsScrobbleReplicationEnabled,
		"isExternalScrobblingEnabled":  settings.IsExternalScrobblingEnabled,
	}

	result := Remote{}
	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, model.ErrNotFound
	} else {
		return result, err
	}
}

func (storage *RemoteStorage) Delete(remoteId uuid.UUID) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		// todo: proper cascading

		if err := tx.Exec("DELETE FROM remote_tracks WHERE remote_id = ?", remoteId).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM remote_albums WHERE remote_id = ?", remoteId).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM remote_artists WHERE remote_id = ?", remoteId).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM remote_artworks WHERE remote_id = ?", remoteId).Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM remote_to_users WHERE remote_id = ?", remoteId).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM remotes WHERE id = ?", remoteId).Error; err != nil {
			return err
		}

		return nil
	})
}

func (storage *RemoteStorage) FindCredentials(userId uuid.UUID, remoteId uuid.UUID) (*RemoteCredentials, error) {
	sql := `
		SELECT *
		FROM remote_to_users
		WHERE user_id = @userId AND remote_id = @remoteId
	`
	params := map[string]any{
		"userId":   userId,
		"remoteId": remoteId,
	}

	result := RemoteToUser{}

	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result.Credentials, err
}

func (storage *RemoteStorage) SetCredentials(userId uuid.UUID, remoteId uuid.UUID, credentials RemoteCredentials) error {
	credentialsText, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO remote_to_users (remote_id, user_id, credentials)
		VALUES (@remoteId, @userId, @credentials)
		ON CONFLICT DO UPDATE
		SET credentials = excluded.credentials
	`
	params := map[string]any{
		"userId":      userId,
		"remoteId":    remoteId,
		"credentials": credentialsText,
	}

	return storage.db.Exec(sql, params).Error
}

func (storage *RemoteStorage) RemoveCredentials(userId uuid.UUID, remoteId uuid.UUID) error {
	sql := `
		DELETE FROM remote_to_users
		WHERE user_id = @userId AND remote_id = @remoteId
	`
	params := map[string]any{
		"userId":   userId,
		"remoteId": remoteId,
	}

	return storage.db.Exec(sql, params).Error
}
