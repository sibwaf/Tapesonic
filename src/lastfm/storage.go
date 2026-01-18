package lastfm

import (
	"errors"
	"tapesonic/model"
	"tapesonic/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LastFmSessionStorage struct {
	db *storage.DbHelper
}

func newLastFmSessionStorage(db *gorm.DB) (*LastFmSessionStorage, error) {
	err := db.AutoMigrate(&LastFmSession{})
	return &LastFmSessionStorage{db: storage.NewDbHelper(db)}, err
}

func (storage *LastFmSessionStorage) Save(session LastFmSession) (LastFmSession, error) {
	sql := `
		INSERT INTO last_fm_sessions (user_id, session_key, username, is_scrobbling_enabled, created_at, updated_at)
		VALUES (@userId, @sessionKey, @username, @isScrobblingEnabled, @createdAt, @updatedAt)
		ON CONFLICT (user_id) DO UPDATE
		SET session_key = excluded.session_key, username = excluded.username, is_scrobbling_enabled = excluded.is_scrobbling_enabled, updated_at = excluded.updated_at
		RETURNING *
	`
	params := map[string]any{
		"userId":              session.UserId,
		"sessionKey":          session.SessionKey,
		"username":            session.Username,
		"isScrobblingEnabled": session.IsScrobblingEnabled,
		"createdAt":           session.CreatedAt,
		"updatedAt":           session.UpdatedAt,
	}

	return session, storage.db.Raw(sql, params).First(&session).Error
}

func (storage *LastFmSessionStorage) UpdateSettings(userId uuid.UUID, settings SessionSettings) (LastFmSession, error) {
	sql := `
		UPDATE last_fm_sessions
		SET is_scrobbling_enabled = @isScrobblingEnabled
		WHERE user_id = @userId
		RETURNING *
	`
	params := map[string]any{
		"userId":              userId,
		"isScrobblingEnabled": settings.IsScrobblingEnabled,
	}

	result := LastFmSession{}
	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, model.ErrNotFound
	} else {
		return result, err
	}
}

func (storage *LastFmSessionStorage) Delete(userId uuid.UUID) error {
	sql := `
		DELETE FROM last_fm_sessions
		WHERE user_id = @userId
	`
	params := map[string]any{
		"userId": userId,
	}

	return storage.db.Exec(sql, params).Error
}

func (storage *LastFmSessionStorage) GetAllSessions() ([]LastFmSession, error) {
	sql := `SELECT * FROM last_fm_sessions ORDER BY user_id`

	result := []LastFmSession{}
	return result, storage.db.Raw(sql).Find(&result).Error
}

func (storage *LastFmSessionStorage) Find(userId uuid.UUID) (*LastFmSession, error) {
	sql := `
		SELECT user_id, session_key, username, is_scrobbling_enabled, created_at, updated_at
		FROM last_fm_sessions
		WHERE user_id = @userId
	`
	params := map[string]any{
		"userId": userId,
	}

	result := LastFmSession{}
	err := storage.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}
