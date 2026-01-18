package listenbrainz

import (
	"errors"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListenBrainzSessionStorage struct {
	db *gorm.DB
}

func newListenBrainzSessionStorage(db *gorm.DB) *ListenBrainzSessionStorage {
	return &ListenBrainzSessionStorage{db: db}
}

func (store *ListenBrainzSessionStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&ListenBrainzSession{})
}

func (store *ListenBrainzSessionStorage) Save(session ListenBrainzSession) (ListenBrainzSession, error) {
	sql := `
		INSERT INTO listen_brainz_sessions (user_id, token, username, is_scrobbling_enabled, created_at, updated_at)
		VALUES (@userId, @token, @username, @isScrobblingEnabled, @createdAt, @updatedAt)
		ON CONFLICT (user_id) DO UPDATE
		SET token = excluded.token, username = excluded.username, is_scrobbling_enabled = excluded.is_scrobbling_enabled, updated_at = excluded.updated_at
		RETURNING *
	`
	params := map[string]any{
		"userId":              session.UserId,
		"token":               session.Token,
		"username":            session.Username,
		"isScrobblingEnabled": session.IsScrobblingEnabled,
		"createdAt":           session.CreatedAt,
		"updatedAt":           session.UpdatedAt,
	}

	return session, store.db.Raw(sql, params).First(&session).Error
}

func (storage *ListenBrainzSessionStorage) UpdateSettings(userId uuid.UUID, settings SessionSettings) (ListenBrainzSession, error) {
	sql := `
		UPDATE listen_brainz_sessions
		SET is_scrobbling_enabled = @isScrobblingEnabled
		WHERE user_id = @userId
		RETURNING *
	`
	params := map[string]any{
		"userId":              userId,
		"isScrobblingEnabled": settings.IsScrobblingEnabled,
	}

	result := ListenBrainzSession{}
	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, model.ErrNotFound
	} else {
		return result, err
	}
}

func (storage *ListenBrainzSessionStorage) Delete(userId uuid.UUID) error {
	sql := `
		DELETE FROM listen_brainz_sessions
		WHERE user_id = @userId
	`
	params := map[string]any{
		"userId": userId,
	}

	return storage.db.Exec(sql, params).Error
}

func (store *ListenBrainzSessionStorage) GetAllSessions() ([]ListenBrainzSession, error) {
	sql := `SELECT * FROM listen_brainz_sessions ORDER BY user_id`

	result := []ListenBrainzSession{}
	return result, store.db.Raw(sql).Find(&result).Error
}

func (store *ListenBrainzSessionStorage) Find(userId uuid.UUID) (*ListenBrainzSession, error) {
	sql := `
		SELECT user_id, token, username, is_scrobbling_enabled, created_at, updated_at
		FROM listen_brainz_sessions
		WHERE user_id = @userId
	`
	params := map[string]any{
		"userId": userId,
	}

	result := ListenBrainzSession{}
	err := store.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}
