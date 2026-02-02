package library

import (
	"fmt"
	"strings"
	"tapesonic/model"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlbumStorage struct {
	db *gorm.DB
}

func newAlbumStorage(db *gorm.DB) *AlbumStorage {
	return &AlbumStorage{db: db}
}

func (store *AlbumStorage) PrepareDatabase() error {
	sql := fmt.Sprintf(
		`
			CREATE VIEW all_albums (
				id,
				artist_id,
				artist,
				title,
				cover_id,
				track_count,
				duration_ms,
				released_at,
				added_at,
				played_at,
				total_listened_ms,
				search_title,
				user_id
			) AS

			WITH
				tapes_aggregate (user_id, id, track_count, total_duration_ms, last_listened_at, total_listened_ms) AS (
					SELECT
						users.id AS user_id,
						tape_to_tracks.tape_id AS id,
						count(*) AS track_count,
						sum(source_tracks.end_offset_ms - source_tracks.start_offset_ms) AS total_duration_ms,
						max(listen_stats.last_listened_at) AS last_listened_at,
						sum((source_tracks.end_offset_ms - source_tracks.start_offset_ms) * listen_stats.listen_count) AS total_listened_ms
					FROM source_tracks
					JOIN users ON 1 = 1
					JOIN tape_to_tracks ON source_tracks.id = tape_to_tracks.track_id
					LEFT JOIN listen_stats ON source_tracks.id = listen_stats.track_id AND users.id = listen_stats.user_id
					GROUP BY users.id, tape_to_tracks.tape_id
				),
				remote_albums_aggregate (user_id, id, track_count, total_duration_ms, last_listened_at, total_listened_ms) AS (
					SELECT
						remote_track_to_users.user_id AS user_id,
						remote_albums.id AS id,
						count(*) AS track_count,
						sum(remote_tracks.duration_ms) AS total_duration_ms,
						max(listen_stats.last_listened_at) AS last_listened_at,
						sum(remote_tracks.duration_ms * listen_stats.listen_count) AS total_listened_ms
					FROM remote_tracks
					JOIN remote_track_to_users ON remote_tracks.id = remote_track_to_users.remote_track_id
					JOIN remote_albums ON remote_tracks.remote_id = remote_albums.remote_id AND remote_tracks.album_id = remote_albums.album_id
					LEFT JOIN listen_stats ON remote_tracks.id = listen_stats.track_id AND remote_track_to_users.user_id = listen_stats.user_id
					GROUP BY remote_track_to_users.user_id, remote_albums.id
				)

			SELECT
				tapes.id AS id,
				NULL AS artist_id,
				tapes.artist AS artist,
				tapes.name AS title,
				tapes.thumbnail_id AS cover_id,
				tapes_aggregate.track_count AS track_count,
				tapes_aggregate.total_duration_ms AS duration_ms,
				tapes.released_at AS released_at,
				tapes.created_at AS added_at,
				tapes_aggregate.last_listened_at AS played_at,
				tapes_aggregate.total_listened_ms AS total_listened_ms,
				tapes.search_name AS search_title,
				users.id AS user_id
			FROM tapes
			JOIN users ON 1 = 1
			LEFT JOIN tapes_aggregate ON tapes.id = tapes_aggregate.id AND users.id = tapes_aggregate.user_id
			WHERE tapes.type = '%s'

			UNION ALL

			SELECT
				remote_albums.id AS id,
				remote_artists.id AS artist_id,
				remote_artists.name AS artist,
				remote_albums.title AS title,
				remote_covers.id AS cover_id,
				remote_albums_aggregate.track_count AS track_count,
				remote_albums_aggregate.total_duration_ms AS duration_ms,
				remote_albums.released_at AS released_at,
				remote_albums.added_at AS added_at,
				remote_albums_aggregate.last_listened_at AS played_at,
				remote_albums_aggregate.total_listened_ms AS total_listened_ms,
				remote_albums.search_title AS search_title,
				remote_album_to_users.user_id AS user_id
			FROM remote_albums
			JOIN remote_album_to_users ON remote_albums.id = remote_album_to_users.remote_album_id
			LEFT JOIN remote_albums_aggregate ON remote_albums.id = remote_albums_aggregate.id AND remote_album_to_users.user_id = remote_albums_aggregate.user_id
			LEFT JOIN remote_covers ON remote_albums.remote_id = remote_covers.remote_id AND remote_albums.cover_id = remote_covers.cover_id
			LEFT JOIN remote_artists ON remote_albums.remote_id = remote_artists.remote_id AND remote_albums.artist_id = remote_artists.artist_id
			WHERE remote_albums_aggregate.track_count > 0
		`,
		model.TAPE_TYPE_ALBUM,
	)

	return store.db.Exec(sql).Error
}

func (store *AlbumStorage) SearchAlbumsByQuery(userId uuid.UUID, query string, count int, offset int) ([]model.LibraryAlbum, error) {
	filter := util.MakeTextSearchCondition([]string{"all_albums.search_title"}, query)
	if filter == "" {
		return []model.LibraryAlbum{}, nil
	}

	return store.getAlbums(userId, count, offset, filter, map[string]any{}, "all_albums.id")
}

func (store *AlbumStorage) GetAlbumById(userId uuid.UUID, albumId string) (model.LibraryAlbum, error) {
	albums, err := store.getAlbums(
		userId,
		1,
		0,
		"all_albums.id = @albumId",
		map[string]any{
			"albumId": albumId,
		},
		"all_albums.id",
	)
	if err != nil {
		return model.LibraryAlbum{}, err
	}
	if len(albums) == 0 {
		return model.LibraryAlbum{}, model.ErrNotFound
	}

	return albums[0], nil
}

func (store *AlbumStorage) GetAlbumsByArtistId(userId uuid.UUID, artistId string) ([]model.LibraryAlbum, error) {
	// todo limit
	return store.getAlbums(
		userId,
		9999,
		0,
		"all_albums.artist_id = @artistId",
		map[string]any{
			"artistId": artistId,
		},
		"all_albums.released_at ASC NULLS FIRST, all_albums.added_at ASC",
	)
}

func (store *AlbumStorage) GetAlbumsSortId(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "", map[string]any{}, "all_albums.id ASC")
}

func (store *AlbumStorage) GetAlbumsSortRandom(userId uuid.UUID, count int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, 0, "", map[string]any{}, "random()")
}

func (store *AlbumStorage) GetAlbumsSortAddedAtDesc(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "", map[string]any{}, "all_albums.added_at DESC")
}

func (store *AlbumStorage) GetAlbumsSortPlayedAtDesc(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "all_albums.played_at IS NOT NULL", map[string]any{}, "all_albums.played_at DESC")
}

func (store *AlbumStorage) GetAlbumsSortTitle(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "", map[string]any{}, "lower(all_albums.title) ASC, id ASC")
}

func (store *AlbumStorage) GetAlbumsSortArtist(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "", map[string]any{}, "lower(all_albums.artist) ASC, all_albums.artist_id, released_at ASC NULLS FIRST, id ASC")
}

func (store *AlbumStorage) GetAlbumsSortTotalListenedDesc(userId uuid.UUID, count int, offset int) ([]model.LibraryAlbum, error) {
	return store.getAlbums(userId, count, offset, "all_albums.total_listened_ms > 0", map[string]any{}, "all_albums.total_listened_ms DESC")
}

func (store *AlbumStorage) GetAlbumsSortReleasedAtDesc(userId uuid.UUID, count int, offset int, fromYear int, toYear int) ([]model.LibraryAlbum, error) {
	var order string
	if fromYear <= toYear {
		order = "all_albums.released_at ASC"
	} else {
		fromYear, toYear = toYear, fromYear
		order = "all_albums.released_at DESC"
	}

	return store.getAlbums(
		userId,
		count,
		offset,
		"all_albums.released_at BETWEEN @minReleasedAt AND @maxReleasedAt",
		map[string]any{
			"minReleasedAt": util.NewTimestampWrapper(time.Date(fromYear, time.January, 1, 0, 0, 0, 0, time.UTC)),
			"maxReleasedAt": util.NewTimestampWrapper(time.Date(toYear+1, time.January, 1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond)),
		},
		order,
	)
}

func (store *AlbumStorage) getAlbums(userId uuid.UUID, count int, offset int, filter string, parameters map[string]any, order string) ([]model.LibraryAlbum, error) {
	query := `
		SELECT
			all_albums.id AS "id",
			all_albums.title AS "name",
			all_albums.artist_id AS "artist_id",
			all_albums.artist AS "artist_name",
			all_albums.cover_id AS "cover_id",
			all_albums.track_count AS "track_count",
			all_albums.duration_ms * 1000000 AS "duration",
			all_albums.released_at AS "released_at",
			all_albums.added_at AS "added_at",
			all_albums.played_at AS "played_at"
			--tapes.updated_at AS "updated_at"
		FROM all_albums
	`

	conditions := []string{
		"all_albums.user_id = @userId",
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

	result := []model.LibraryAlbum{}
	if len(allParameters) > 0 {
		return result, store.db.Raw(query, allParameters).Find(&result).Error
	} else {
		return result, store.db.Raw(query).Find(&result).Error
	}
}
