package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaylistStorage struct {
	db *gorm.DB
}

func NewPlaylistStorage(db *gorm.DB) (*PlaylistStorage, error) {
	err := db.AutoMigrate()
	return &PlaylistStorage{db: db}, err
}

func (storage *PlaylistStorage) GetSubsonicPlaylist(id uuid.UUID) (*SubsonicPlaylistItem, error) {
	playlists, err := storage.getSubsonicPlaylists(1, 0, fmt.Sprintf("tapes.id = '%s'", id.String()), "tapes.id")
	if err != nil {
		return nil, err
	}
	if len(playlists) == 0 {
		return nil, fmt.Errorf("playlist with id %s doesn't exist", id.String())
	}

	return &playlists[0], nil
}

func (storage *PlaylistStorage) GetSubsonicPlaylists(count int, offset int) ([]SubsonicPlaylistItem, error) {
	return storage.getSubsonicPlaylists(count, offset, "", "tapes.updated_at DESC")
}

func (storage *PlaylistStorage) getSubsonicPlaylists(count int, offset int, filter string, order string) ([]SubsonicPlaylistItem, error) {
	query := `
		WITH playlist_extra_info AS (
			SELECT
				tape_to_tracks.tape_id AS tape_id,
				count(*) AS song_count,
				sum(tracks.end_offset_ms - tracks.start_offset_ms) / 1000 AS duration_sec
			FROM tape_to_tracks
			LEFT JOIN tracks ON tracks.id = tape_to_tracks.track_id
			GROUP BY tape_to_tracks.tape_id
		)
		SELECT *
		FROM tapes
		LEFT JOIN playlist_extra_info ON playlist_extra_info.tape_id = tapes.id
	`

	conditions := []string{
		fmt.Sprintf("tapes.type = '%s'", TAPE_TYPE_PLAYLIST),
	}

	if filter != "" {
		conditions = append(conditions, filter)
	}

	query += fmt.Sprintf("\nWHERE %s", strings.Join(conditions, " AND "))
	query += fmt.Sprintf("\nORDER BY %s", order)
	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	result := []SubsonicPlaylistItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}
