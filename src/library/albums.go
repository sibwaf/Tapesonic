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
			all_albums.artwork_id AS "artwork_id",
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
