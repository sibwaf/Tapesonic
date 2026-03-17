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

func (store *TrackStorage) PrepareDatabase() error {
	sql := `
		CREATE TABLE IF NOT EXISTS all_track_ids (
			id TEXT NOT NULL PRIMARY KEY,
			source_track_id TEXT,
			remote_track_id TEXT,

			CONSTRAINT all_track_ids_source_track_id_fk FOREIGN KEY (source_track_id) REFERENCES source_tracks (id),
			CONSTRAINT all_track_ids_remote_track_id_fk FOREIGN KEY (remote_track_id) REFERENCES remote_tracks (id)
		)
	`
	if err := store.db.Exec(sql).Error; err != nil {
		return err
	}

	sql = fmt.Sprintf(
		`
			CREATE VIEW all_tracks (
				id,
				source_id,
				remote_id,
				remote_track_id,
				artist,
				artist_id,
				album,
				album_id,
				album_released_at,
				album_track_index,
				title,
				cover_id,
				duration_ms,
				played_at,
				search_artist,
				search_album,
				search_title,
				user_id
			) AS

			WITH track_albums (track_id, tape_id, list_index) AS (
				SELECT
					tape_to_tracks.track_id AS track_id,
					tapes.id AS tape_id,
					tape_to_tracks.list_index AS list_index
				FROM tape_to_tracks
				JOIN tapes ON tape_to_tracks.tape_id = tapes.id
				WHERE tapes.type = '%s'
			)

			SELECT
				source_tracks.id AS id,
				sources.id AS source_id,
				NULL AS remote_id,
				NULL AS remote_track_id,
				artists.name AS artist,
				artists.id AS artist_id,
				tapes.name AS album,
				tapes.id AS album_id,
				tapes.released_at AS album_released_at,
				track_albums.list_index AS album_track_index,
				source_tracks.title AS title,
				coalesce(tapes.thumbnail_id, sources.thumbnail_id) AS cover_id,
				source_tracks.end_offset_ms - source_tracks.start_offset_ms AS duration_ms,
				listen_stats.last_listened_at AS played_at,
				artists.search_name AS search_artist,
				tapes.search_name AS search_album,
				source_tracks.search_title AS search_title,
				users.id AS user_id
			FROM source_tracks
			JOIN users ON 1 = 1
			LEFT JOIN artists ON source_tracks.artist_id = artists.id
			LEFT JOIN track_albums ON source_tracks.id = track_albums.track_id
			LEFT JOIN tapes ON track_albums.tape_id = tapes.id
			LEFT JOIN sources ON source_tracks.source_id = sources.id
			LEFT JOIN listen_stats ON source_tracks.id = listen_stats.track_id AND users.id = listen_stats.user_id

			UNION ALL

			SELECT
				remote_tracks.id AS id,
				NULL AS source_id,
				remote_tracks.remote_id AS remote_id,
				remote_tracks.track_id AS remote_track_id,
				artists.name AS artist,
				artists.id AS artist_id,
				remote_albums.title AS album,
				remote_albums.id AS album_id,
				remote_albums.released_at AS album_released_at,
				remote_tracks.album_index AS album_track_index,
				remote_tracks.title AS title,
				remote_covers.id AS cover_id,
				remote_tracks.duration_ms AS duration_ms,
				listen_stats.last_listened_at AS played_at,
				artists.search_name AS search_artist,
				remote_albums.search_title AS search_album,
				remote_tracks.search_title AS search_title,
				remote_track_to_users.user_id AS user_id
			FROM remote_tracks
			JOIN remote_track_to_users ON remote_tracks.id = remote_track_to_users.remote_track_id
			LEFT JOIN remote_artists ON remote_tracks.remote_id = remote_artists.remote_id AND remote_tracks.artist_id = remote_artists.artist_id
			LEFT JOIN artists ON remote_artists.tapesonic_artist_id = artists.id
			LEFT JOIN remote_albums ON remote_tracks.remote_id = remote_albums.remote_id AND remote_tracks.album_id = remote_albums.album_id
			LEFT JOIN remote_covers ON remote_tracks.remote_id = remote_covers.remote_id AND remote_tracks.cover_id = remote_covers.cover_id
			LEFT JOIN listen_stats ON remote_tracks.id = listen_stats.track_id AND remote_track_to_users.user_id = listen_stats.user_id
		`,
		model.TAPE_TYPE_ALBUM,
	)
	if err := store.db.Exec(sql).Error; err != nil {
		return err
	}

	return nil
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
		len(ids),
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
			all_tracks.album AS "album_name",
			all_tracks.album_track_index AS "album_track_index",
			all_tracks.cover_id AS "cover_id",
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
