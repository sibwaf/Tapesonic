package library

import (
	"fmt"
	"slices"
	"strings"
	"tapesonic/model"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackStorage struct {
	db *gorm.DB
}

func newTrackStorage(db *gorm.DB) *TrackStorage {
	return &TrackStorage{db: db}
}

func (store *TrackStorage) SearchTracksByQuery(userId uuid.UUID, query string, count int, offset int) ([]model.LibraryTrack, error) {
	filter := util.MakeTextSearchCondition([]string{"all_tracks.search_artist", "all_tracks.search_title", "all_tracks.search_album"}, query)
	if filter == "" {
		return []model.LibraryTrack{}, nil
	}

	return store.getTracks(userId, count, offset, filter, map[string]any{}, "all_tracks.id")
}

func (store *TrackStorage) SearchTracksByFields(userId uuid.UUID, filter TrackFilter, count int, offset int) ([]model.LibraryTrack, error) {
	conditions := []string{
		util.MakeTextSearchCondition([]string{"all_tracks.search_artist"}, filter.Artist),
		util.MakeTextSearchCondition([]string{"all_tracks.search_album"}, filter.Album),
		util.MakeTextSearchCondition([]string{"all_tracks.search_title"}, filter.Title),
	}
	conditions = slices.DeleteFunc(conditions, func(condition string) bool { return condition == "" })

	if len(conditions) == 0 {
		return []model.LibraryTrack{}, nil
	}

	return store.getTracks(userId, count, offset, strings.Join(conditions, " AND "), map[string]any{}, "all_tracks.id")
}

func (store *TrackStorage) GetTrackById(userId uuid.UUID, id string) (model.LibraryTrack, error) {
	tracks, err := store.getTracks(
		userId,
		1,
		0,
		"all_tracks.id = @trackId",
		map[string]any{"trackId": id},
		"all_tracks.id",
	)
	if err != nil {
		return model.LibraryTrack{}, err
	}
	if len(tracks) == 0 {
		return model.LibraryTrack{}, model.ErrNotFound
	}

	return tracks[0], nil
}

func (store *TrackStorage) GetTracksByIds(userId uuid.UUID, ids []string) ([]model.LibraryTrack, error) {
	if len(ids) == 0 {
		return []model.LibraryTrack{}, nil
	}

	return store.getTracks(
		userId,
		9999, // todo: set to len(ids) when track ids stop duplicating due to multiple albums
		0,
		"all_tracks.id IN @trackIds",
		map[string]any{"trackIds": ids},
		"all_tracks.id",
	)
}

func (store *TrackStorage) GetTracksByAlbumId(userId uuid.UUID, albumId string) ([]model.LibraryTrack, error) {
	// todo limit
	return store.getTracks(
		userId,
		9999,
		0,
		"all_tracks.album_id = @albumId",
		map[string]any{"albumId": albumId},
		"all_tracks.album_track_index",
	)
}

func (store *TrackStorage) GetTracksSortId(userId uuid.UUID, count int, offset int) ([]model.LibraryTrack, error) {
	return store.getTracks(userId, count, offset, "", map[string]any{}, "all_tracks.id ASC")
}

func (store *TrackStorage) GetTracksSortRandom(userId uuid.UUID, count int, fromYear *int, toYear *int) ([]model.LibraryTrack, error) {
	conditions := []string{}
	params := map[string]any{}

	if fromYear != nil {
		conditions = append(conditions, "all_tracks.album_released_at >= @minReleasedAt")
		params["minReleasedAt"] = util.NewTimestampWrapper(time.Date(*fromYear, time.January, 1, 0, 0, 0, 0, time.UTC))
	}
	if toYear != nil {
		conditions = append(conditions, "all_tracks.album_released_at <= @maxReleasedAt")
		params["maxReleasedAt"] = util.NewTimestampWrapper(time.Date(*toYear+1, time.January, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond))
	}

	return store.getTracks(
		userId,
		count,
		0,
		strings.Join(conditions, " AND "),
		params,
		"random()",
	)
}

func (store *TrackStorage) getTracks(userId uuid.UUID, count int, offset int, filter string, parameters map[string]any, order string) ([]model.LibraryTrack, error) {
	query := `
		SELECT
			all_tracks.id AS "id",
			all_tracks.source_id AS "source_id",
			all_tracks.remote_id AS "remote_id",
			all_tracks.remote_track_id AS "remote_track_id",
			all_tracks.title AS "title",
			all_tracks.artist_id AS "artist_id",
			all_tracks.artist AS "artist_name",
			all_tracks.album_id AS "album_id",
			all_tracks.album_artist_id AS "album_artist_id",
			all_tracks.album_released_at AS "album_released_at",
			all_tracks.album AS "album_name",
			all_tracks.album_track_index AS "album_track_index",
			all_tracks.artwork_id AS "artwork_id",
			all_tracks.duration_ms * 1000 * 1000 AS "duration",
			all_tracks.played_at AS "played_at"
		FROM all_tracks
	`

	conditions := []string{
		"all_tracks.user_id = @userId",
	}
	if filter != "" {
		conditions = append(conditions, filter)
	}
	if len(conditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s", strings.Join(conditions, " AND "))
	}

	if order != "" {
		query += fmt.Sprintf("\nORDER BY %s", order)
	}

	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	allParameters := map[string]any{}
	for key, value := range parameters {
		allParameters[key] = value
	}
	allParameters["userId"] = userId

	result := []model.LibraryTrack{}
	if len(allParameters) > 0 {
		return result, store.db.Raw(query, allParameters).Find(&result).Error
	} else {
		return result, store.db.Raw(query).Find(&result).Error
	}
}
