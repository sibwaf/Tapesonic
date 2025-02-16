package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ListenbrainzPlaylist struct {
	Id          string
	Name        string
	Description string
	CreatedBy   string

	CreatedAt time.Time
	UpdatedAt time.Time

	Tracks []*ListenbrainzPlaylistTrack
}

type ListenbrainzPlaylistTrack struct {
	Artist string
	Album  string
	Title  string

	ListenbrainzPlaylistId string
	ListenbrainzPlaylist   *ListenbrainzPlaylist

	MatchedServiceName string
	MatchedSongId      string

	TrackIndex int
}

type ListenbrainzPlaylistStorage struct {
	db *DbHelper
}

func NewListenBrainzPlaylistStorage(db *gorm.DB) (*ListenbrainzPlaylistStorage, error) {
	if err := db.AutoMigrate(&ListenbrainzPlaylist{}, &ListenbrainzPlaylistTrack{}); err != nil {
		return nil, err
	}

	return &ListenbrainzPlaylistStorage{
		db: NewDbHelper(db),
	}, nil
}

func (storage *ListenbrainzPlaylistStorage) Replace(items []ListenbrainzPlaylist) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ListenbrainzPlaylistTrack{}).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ListenbrainzPlaylist{}).Error; err != nil {
			return err
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (storage *ListenbrainzPlaylistStorage) GetSubsonicPlaylist(id string) (*SubsonicPlaylistItem, error) {
	playlists, err := storage.getSubsonicPlaylists(1, 0, fmt.Sprintf("listenbrainz_playlists.id = '%s'", id), "listenbrainz_playlists.id")
	if err != nil {
		return nil, err
	}
	if len(playlists) == 0 {
		return nil, fmt.Errorf("listenbrainz playlist with id %s doesn't exist", id)
	}

	return &playlists[0], nil
}

func (storage *ListenbrainzPlaylistStorage) GetSubsonicPlaylists(count int, offset int) ([]SubsonicPlaylistItem, error) {
	return storage.getSubsonicPlaylists(count, offset, "", "listenbrainz_playlists.created_at ASC")
}

func (storage *ListenbrainzPlaylistStorage) getSubsonicPlaylists(count int, offset int, filter string, order string) ([]SubsonicPlaylistItem, error) {
	query := `
		WITH playlist_extra_info AS (
			SELECT
				listenbrainz_playlist_tracks.listenbrainz_playlist_id AS playlist_id,
				count(*) AS song_count,
				sum(cached_mux_songs.duration_sec) AS duration_sec
			FROM listenbrainz_playlist_tracks
			JOIN cached_mux_songs ON cached_mux_songs.service_name = listenbrainz_playlist_tracks.matched_service_name AND cached_mux_songs.song_id = listenbrainz_playlist_tracks.matched_song_id
			GROUP BY listenbrainz_playlist_tracks.listenbrainz_playlist_id
		)
		SELECT *
		FROM listenbrainz_playlists
		LEFT JOIN playlist_extra_info ON playlist_extra_info.playlist_id = listenbrainz_playlists.id
	`

	if filter != "" {
		query += fmt.Sprintf("\nWHERE %s", filter)
	}

	query += fmt.Sprintf("\nORDER BY %s", order)
	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	result := []SubsonicPlaylistItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *ListenbrainzPlaylistStorage) GetTracksByPlaylist(playlistId string) ([]CachedSongIdWithIndex, error) {
	result := []CachedSongIdWithIndex{}

	query := `
		SELECT
			listenbrainz_playlist_tracks.matched_service_name AS service_name,
			listenbrainz_playlist_tracks.matched_song_id AS id,
			listenbrainz_playlist_tracks.track_index AS track_index
		FROM listenbrainz_playlist_tracks
		JOIN cached_mux_songs ON cached_mux_songs.service_name = listenbrainz_playlist_tracks.matched_service_name AND cached_mux_songs.song_id = listenbrainz_playlist_tracks.matched_song_id
		WHERE listenbrainz_playlist_tracks.listenbrainz_playlist_id = ?
		ORDER BY listenbrainz_playlist_tracks.track_index
	`

	return result, storage.db.Raw(query, playlistId).Find(&result).Error
}
