package recommendations

import (
	"errors"
	"tapesonic/model"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecommendationStorage struct {
	db *gorm.DB
}

func newRecommendationStorage(db *gorm.DB) *RecommendationStorage {
	return &RecommendationStorage{db: db}
}

func (store *RecommendationStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&RecommendedPlaylist{}, &RecommendedPlaylistTrack{})
}

func (store *RecommendationStorage) GetById(id uuid.UUID) (RecommendedPlaylist, error) {
	sql := `SELECT * FROM recommended_playlists WHERE id = @id`
	params := map[string]any{"id": id}

	result := RecommendedPlaylist{}
	err := store.db.Raw(sql, params).First(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, model.ErrNotFound
	} else {
		return result, err
	}
}

func (store *RecommendationStorage) UpsertPlaylist(playlist RecommendedPlaylist) (RecommendedPlaylist, error) {
	sql := `
		INSERT INTO recommended_playlists (id, provider, provider_playlist_id, user_id, name, cover_id, created_at, updated_at, sync_tag)
		VALUES (@id, @provider, @providerPlaylistId, @userId, @name, @coverId, @createdAt, @updatedAt, @syncTag)
		ON CONFLICT (provider, provider_playlist_id) DO UPDATE
		SET name = excluded.name, cover_id = excluded.cover_id, sync_tag = excluded.sync_tag
		RETURNING *
	`
	params := map[string]any{
		"id":                 playlist.Id,
		"provider":           playlist.Provider,
		"providerPlaylistId": playlist.ProviderPlaylistId,
		"userId":             playlist.UserId,
		"name":               playlist.Name,
		"coverId":            playlist.CoverId,
		"createdAt":          playlist.CreatedAt,
		"updatedAt":          playlist.UpdatedAt,
		"syncTag":            playlist.SyncTag,
	}

	return playlist, store.db.Raw(sql, params).First(&playlist).Error
}

func (store *RecommendationStorage) ReplaceTracks(playlistId uuid.UUID, tracks []RecommendedPlaylistTrack) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM recommended_playlist_tracks WHERE recommended_playlist_id = @playlistId", map[string]any{"playlistId": playlistId}).Error; err != nil {
			return err
		}

		sql := `
			INSERT INTO recommended_playlist_tracks (recommended_playlist_id, artist, title, track_id, track_index)
			VALUES (@recommendedPlaylistId, @artist, @title, @trackId, @trackIndex)
		`
		for i, track := range tracks {
			params := map[string]any{
				"recommendedPlaylistId": playlistId,
				"artist":                track.Artist,
				"title":                 track.Title,
				"trackId":               track.TrackId,
				"trackIndex":            i,
			}
			if err := tx.Exec(sql, params).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (store *RecommendationStorage) DeleteUnsyncedPlaylists(provider RecommendationProvider, userId uuid.UUID, syncTag string) error {
	sql := `
		DELETE FROM recommended_playlists
		WHERE 1 = 1
			AND recommended_playlists.provider = @provider
			AND recommended_playlists.user_id = @userId
			AND recommended_playlists.sync_tag != @syncTag
	`
	params := map[string]any{
		"provider": provider,
		"userId":   userId,
		"syncTag":  syncTag,
	}
	if err := store.db.Exec(sql, params).Error; err != nil {
		return err
	}

	return nil
}

func (store *RecommendationStorage) UpdatePlaylistUpdatedAt(id uuid.UUID, updatedAt time.Time) error {
	sql := `
		UPDATE recommended_playlists
		SET updated_at = @updatedAt
		WHERE id = @id
	`
	params := map[string]any{
		"id":        id,
		"updatedAt": util.NewTimestampWrapper(updatedAt),
	}
	return store.db.Exec(sql, params).Error
}
