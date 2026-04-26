package library

import (
	"errors"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaylistStorage struct {
	db *gorm.DB
}

func newPlaylistStorage(db *gorm.DB) *PlaylistStorage {
	return &PlaylistStorage{db: db}
}

func (store *PlaylistStorage) GetAllPlaylists(userId uuid.UUID) ([]model.LibraryPlaylist, error) {
	sql := `
		SELECT
			all_playlists.id AS id,
			all_playlists.name AS name,
			all_playlists.artwork_id AS artwork_id,
			all_playlists.track_count AS track_count,
			all_playlists.duration_ms * 1000 * 1000 AS duration,
			all_playlists.created_at AS created_at,
			all_playlists.updated_at AS updated_at
		FROM all_playlists
		WHERE all_playlists.user_id = @userId
		ORDER BY all_playlists.updated_at DESC
	`
	params := map[string]any{"userId": userId}

	result := []model.LibraryPlaylist{}
	return result, store.db.Raw(sql, params).Find(&result).Error
}

func (store *PlaylistStorage) GetPlaylistById(userId uuid.UUID, playlistId string) (model.LibraryPlaylist, error) {
	sql := `
		SELECT
			all_playlists.id AS id,
			all_playlists.name AS name,
			all_playlists.artwork_id AS artwork_id,
			all_playlists.track_count AS track_count,
			all_playlists.duration_ms * 1000 * 1000 AS duration,
			all_playlists.created_at AS created_at,
			all_playlists.updated_at AS updated_at
		FROM all_playlists
		WHERE all_playlists.user_id = @userId AND all_playlists.id = @playlistId
	`
	params := map[string]any{
		"userId":     userId,
		"playlistId": playlistId,
	}

	result := model.LibraryPlaylist{}
	err := store.db.Raw(sql, params).Find(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = model.ErrNotFound
	}

	return result, err
}

func (store *PlaylistStorage) GetTrackIdsByPlaylistId(playlistId string) ([]uuid.UUID, error) {
	sql := `
		SELECT all_playlist_tracks.track_id
		FROM all_playlist_tracks
		WHERE all_playlist_tracks.playlist_id = @playlistId
		ORDER BY all_playlist_tracks.track_index ASC
	`
	params := map[string]any{"playlistId": playlistId}

	result := []uuid.UUID{}
	return result, store.db.Raw(sql, params).Find(&result).Error
}
