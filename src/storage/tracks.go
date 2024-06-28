package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxCount = 9999
)

type TrackStorage struct {
	db *DbHelper
}

func NewTrackStorage(db *gorm.DB) (*TrackStorage, error) {
	return &TrackStorage{db: NewDbHelper(db)}, nil
}

func (storage *TrackStorage) GetSubsonicTrack(id uuid.UUID) (*SubsonicTrackItem, error) {
	tracks, err := storage.getSubsonicTracks(1, 0, fmt.Sprintf("tape_tracks.id = '%s'", id.String()), "", "tape_tracks.id")
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("track with id %s doesn't exist", id.String())
	}

	return &tracks[0], nil
}

func (storage *TrackStorage) SearchSubsonicTracks(count int, offset int, query []string) ([]SubsonicTrackItem, error) {
	filter := []string{}
	for _, term := range query {
		searchField := "' ' || tape_tracks.artist || ' ' || coalesce(albums.name, '') || ' ' || tape_tracks.title"
		filter = append(filter, fmt.Sprintf("%s LIKE '%% %s%%' ESCAPE '%s'", searchField, EscapeTextLiteralForLike(term, "\\"), "\\"))
	}

	return storage.getSubsonicTracks(count, offset, strings.Join(filter, " AND "), "", "tape_tracks.id")
}

func (storage *TrackStorage) GetSubsonicTracksSortId(count int, offset int) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks(count, offset, "", "", "tape_tracks.id")
}

func (storage *TrackStorage) GetSubsonicTracksSortRandom(count int, fromYear *int, toYear *int) ([]SubsonicTrackItem, error) {
	conditions := []string{}

	if fromYear != nil && toYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', albums.release_date) AS INTEGER) BETWEEN %d AND %d", *fromYear, *toYear))
	} else if fromYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', albums.release_date) AS INTEGER) >= %d", *fromYear))
	} else if toYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', albums.release_date) AS INTEGER) <= %d", *toYear))
	}

	return storage.getSubsonicTracks(count, 0, strings.Join(conditions, " AND "), "", "random()")
}

func (storage *TrackStorage) GetSubsonicTracksByAlbum(albumId uuid.UUID) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks(maxCount, 0, fmt.Sprintf("album_tracks.album_id = '%s'", albumId.String()), "", "album_tracks.track_index")
}

func (storage *TrackStorage) GetSubsonicTracksByPlaylist(playlistId uuid.UUID) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks(maxCount, 0, "", playlistId.String(), "playlist_tracks.track_index")
}

func (storage *TrackStorage) getSubsonicTracks(
	count int,
	offset int,
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

	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	result := []SubsonicTrackItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}
