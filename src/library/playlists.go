package library

import (
	"errors"
	"fmt"
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

func (store *PlaylistStorage) PrepareDatabase() error {
	sqlPlaylists := fmt.Sprintf(
		`
			CREATE VIEW all_playlists (
				id,
				name,
				cover_id,
				track_count,
				duration_ms,
				created_at,
				updated_at,
				user_id
			) AS

			WITH
				tapes_aggregate (user_id, id, track_count, total_duration_ms) AS (
					SELECT
						users.id AS user_id,
						tape_to_tracks.tape_id AS id,
						count(*) AS track_count,
						sum(source_tracks.end_offset_ms - source_tracks.start_offset_ms) AS total_duration_ms
					FROM source_tracks
					JOIN users ON 1 = 1
					JOIN tape_to_tracks ON source_tracks.id = tape_to_tracks.track_id
					GROUP BY users.id, tape_to_tracks.tape_id
				),
				recommended_playlists_aggregate (user_id, id, track_count, total_duration_ms) AS (
					SELECT
						all_tracks.user_id AS user_id,
						recommended_playlist_tracks.recommended_playlist_id AS id,
						count(*) AS track_count,
						sum(all_tracks.duration_ms) AS total_duration_ms
					FROM recommended_playlist_tracks
					JOIN all_tracks ON recommended_playlist_tracks.track_id = all_tracks.id
					GROUP BY recommended_playlist_tracks.recommended_playlist_id, all_tracks.user_id
				)

			SELECT
				tapes.id AS id,
				tapes.name AS name,
				tapes.thumbnail_id AS cover_id,
				tapes_aggregate.track_count AS track_count,
				tapes_aggregate.total_duration_ms AS duration_ms,
				tapes.created_at AS created_at,
				tapes.updated_at AS updated_at,
				users.id AS user_id
			FROM tapes
			JOIN users ON 1 = 1
			LEFT JOIN tapes_aggregate ON tapes.id = tapes_aggregate.id AND users.id = tapes_aggregate.user_id
			WHERE tapes.type = '%s'

			UNION ALL

			SELECT
				recommended_playlists.id AS id,
				recommended_playlists.name AS name,
				NULL AS cover_id,
				recommended_playlists_aggregate.track_count AS track_count,
				recommended_playlists_aggregate.total_duration_ms AS duration_ms,
				recommended_playlists.created_at AS created_at,
				recommended_playlists.updated_at AS updated_at,
				recommended_playlists.user_id AS user_id
			FROM recommended_playlists
			LEFT JOIN recommended_playlists_aggregate ON recommended_playlists.id = recommended_playlists_aggregate.id AND recommended_playlists.user_id = recommended_playlists_aggregate.user_id
			WHERE recommended_playlists_aggregate.track_count > 0
		`,
		model.TAPE_TYPE_PLAYLIST,
	)
	if err := store.db.Exec(sqlPlaylists).Error; err != nil {
		return err
	}

	sqlTracks := fmt.Sprintf(
		`
			CREATE VIEW all_playlist_tracks (
				playlist_id,
				track_id,
				track_index
			) AS

			SELECT
				tape_to_tracks.tape_id AS playlist_id,
				tape_to_tracks.track_id AS track_id,
				tape_to_tracks.list_index AS track_index
			FROM tape_to_tracks
			JOIN tapes ON tape_to_tracks.tape_id = tapes.id
			WHERE tapes.type = '%s'

			UNION ALL

			SELECT
				recommended_playlist_tracks.recommended_playlist_id AS playlist_id,
				recommended_playlist_tracks.track_id AS track_id,
				recommended_playlist_tracks.track_index AS track_index
			FROM recommended_playlist_tracks
		`,
		model.TAPE_TYPE_PLAYLIST,
	)
	if err := store.db.Exec(sqlTracks).Error; err != nil {
		return err
	}

	return nil
}

func (store *PlaylistStorage) GetAllPlaylists(userId uuid.UUID) ([]model.LibraryPlaylist, error) {
	sql := `
		SELECT
			all_playlists.id AS id,
			all_playlists.name AS name,
			all_playlists.cover_id AS cover_id,
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
			all_playlists.cover_id AS cover_id,
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
