package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExternalPlaylist struct {
	Id string

	Provider string `gorm:"uniqueIndex:uniq"`
	RawId    string `gorm:"uniqueIndex:uniq"`

	Name        string
	Description string
	CreatedBy   string

	CreatedAt time.Time
	UpdatedAt time.Time

	Tracks []ExternalPlaylistTrack
}

type ExternalPlaylistTrack struct {
	Artist string
	Album  string
	Title  string

	ExternalPlaylistId string
	ExternalPlaylist   *ExternalPlaylist

	MatchedServiceName string
	MatchedSongId      string

	TrackIndex int
}

type ExternalPlaylistStorage struct {
	db *DbHelper
}

func NewExternalPlaylistStorage(db *gorm.DB) (*ExternalPlaylistStorage, error) {
	if err := db.AutoMigrate(&ExternalPlaylist{}, &ExternalPlaylistTrack{}); err != nil {
		return nil, err
	}

	return &ExternalPlaylistStorage{
		db: NewDbHelper(db),
	}, nil
}

func (storage *ExternalPlaylistStorage) Replace(provider string, playlists []ExternalPlaylist) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`
				DELETE FROM external_playlist_tracks
				WHERE EXISTS (
					SELECT 1 FROM external_playlists
					WHERE external_playlists.id = external_playlist_tracks.external_playlist_id AND external_playlists.provider = ?
				)
			`,
			provider,
		).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			`
				DELETE FROM external_playlists
				WHERE external_playlists.provider = ?
			`,
			provider,
		).Error; err != nil {
			return err
		}

		if len(playlists) > 0 {
			for i := range playlists {
				playlists[i].Provider = provider
			}

			if err := tx.Omit(clause.Associations).CreateInBatches(&playlists, 256).Error; err != nil {
				return err
			}
		}

		tracks := []ExternalPlaylistTrack{}
		for _, playlist := range playlists {
			for i := range playlist.Tracks {
				playlist.Tracks[i].ExternalPlaylistId = playlist.Id
			}
			tracks = append(tracks, playlist.Tracks...)
		}

		if len(tracks) > 0 {
			if err := tx.Omit(clause.Associations).CreateInBatches(&tracks, 256).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (storage *ExternalPlaylistStorage) GetSubsonicPlaylist(id string) (SubsonicPlaylistItem, error) {
	playlists, err := storage.getSubsonicPlaylists(1, 0, fmt.Sprintf("external_playlists.id = '%s'", id), "external_playlists.id")
	if err != nil {
		return SubsonicPlaylistItem{}, err
	}
	if len(playlists) == 0 {
		return SubsonicPlaylistItem{}, fmt.Errorf("external playlist with id %s doesn't exist", id)
	}

	return playlists[0], nil
}

func (storage *ExternalPlaylistStorage) GetSubsonicPlaylists(count int, offset int) ([]SubsonicPlaylistItem, error) {
	return storage.getSubsonicPlaylists(count, offset, "", "external_playlists.created_at ASC")
}

func (storage *ExternalPlaylistStorage) getSubsonicPlaylists(count int, offset int, filter string, order string) ([]SubsonicPlaylistItem, error) {
	query := `
		WITH playlist_extra_info AS (
			SELECT
				external_playlist_tracks.external_playlist_id AS playlist_id,
				count(*) AS song_count,
				sum(cached_mux_songs.duration_sec) AS duration_sec
			FROM external_playlist_tracks
			JOIN cached_mux_songs ON cached_mux_songs.service_name = external_playlist_tracks.matched_service_name AND cached_mux_songs.song_id = external_playlist_tracks.matched_song_id
			GROUP BY external_playlist_tracks.external_playlist_id
		)
		SELECT *
		FROM external_playlists
		LEFT JOIN playlist_extra_info ON playlist_extra_info.playlist_id = external_playlists.id
	`

	if filter != "" {
		query += fmt.Sprintf("\nWHERE %s", filter)
	}

	query += fmt.Sprintf("\nORDER BY %s", order)
	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	result := []SubsonicPlaylistItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *ExternalPlaylistStorage) GetTracksByPlaylist(playlistId string) ([]CachedSongIdWithIndex, error) {
	result := []CachedSongIdWithIndex{}

	query := `
		SELECT
			external_playlist_tracks.matched_service_name AS service_name,
			external_playlist_tracks.matched_song_id AS id,
			external_playlist_tracks.track_index AS track_index
		FROM external_playlist_tracks
		JOIN cached_mux_songs ON cached_mux_songs.service_name = external_playlist_tracks.matched_service_name AND cached_mux_songs.song_id = external_playlist_tracks.matched_song_id
		WHERE external_playlist_tracks.external_playlist_id = ?
		ORDER BY external_playlist_tracks.track_index
	`

	return result, storage.db.Raw(query, playlistId).Find(&result).Error
}
