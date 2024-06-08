package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackStorage struct {
	db *DbHelper
}

func NewTrackStorage(db *gorm.DB) (*TrackStorage, error) {
	return &TrackStorage{db: NewDbHelper(db)}, nil
}

func (storage *TrackStorage) GetSubsonicTrack(id uuid.UUID) (*SubsonicTrackItem, error) {
	tracks, err := storage.getSubsonicTracks(fmt.Sprintf("tape_tracks.id = '%s'", id.String()), "", "tape_tracks.id")
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("track with id %s doesn't exist", id.String())
	}

	return &tracks[0], nil
}

func (storage *TrackStorage) GetSubsonicTracksByAlbum(albumId uuid.UUID) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks(fmt.Sprintf("album_tracks.album_id = '%s'", albumId.String()), "", "album_tracks.track_index")
}

func (storage *TrackStorage) GetSubsonicTracksByPlaylist(playlistId uuid.UUID) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks("", playlistId.String(), "playlist_tracks.track_index")
}

func (storage *TrackStorage) getSubsonicTracks(
	filter string,
	playlistId string,
	order string,
) ([]SubsonicTrackItem, error) {
	fields := []string{
		"tape_tracks.*",
		"albums.id AS album_id",
		"albums.name AS album",
		"album_tracks.track_index AS album_track_index",
		"(tape_tracks.end_offset_ms - tape_tracks.start_offset_ms) / 1000 AS duration_sec",
		"tape_tracks.artist AS artist",
		"tape_tracks.title AS title",
		"tape_track_listens.listen_count AS play_count",
	}
	joins := []string{
		"album_tracks ON album_tracks.tape_track_id = tape_tracks.id",
		"albums ON albums.id = album_tracks.album_id",

		"tape_track_listens ON tape_track_listens.tape_track_id = tape_tracks.id",
	}
	conditions := []string{}

	if filter != "" {
		conditions = append(conditions, filter)
	}

	if playlistId != "" {
		fields = append(fields, "playlist_tracks.track_index AS playlist_track_index")
		joins = append(
			joins,
			"playlist_tracks ON playlist_tracks.tape_track_id = tape_tracks.id",
			"playlists ON playlists.id = playlist_tracks.playlist_id",
		)
		conditions = append(conditions, fmt.Sprintf("playlists.id = '%s'", playlistId))
	}

	query := fmt.Sprintf("SELECT %s FROM tape_tracks", strings.Join(fields, ", "))

	for _, join := range joins {
		query += fmt.Sprintf("\nLEFT JOIN %s", join)
	}

	if len(conditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s", strings.Join(conditions, " AND "))
	}

	query += fmt.Sprintf("\nORDER BY %s", order)

	result := []SubsonicTrackItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}
