package library

import (
	"errors"
	"fmt"
	"tapesonic/model"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtistStorage struct {
	db *gorm.DB
}

func newArtistStorage(db *gorm.DB) *ArtistStorage {
	return &ArtistStorage{db: db}
}

func (store *ArtistStorage) PrepareDatabase() error {
	sql := `
		CREATE VIEW all_artists (
			id,
			name,
			cover_id,
			album_count,
			search_name,
			user_id
		) AS

		WITH
			artist_raw_aggregate (user_id, artist_id, album_count) AS (
				SELECT
					users.id AS user_id,
					tapes.artist_id AS artist_id,
					count(*) AS album_count
				FROM tapes
				JOIN users ON 1 = 1
				WHERE tapes."type" = 'album'
				GROUP BY users.id, tapes.artist_id

				UNION ALL

				SELECT
					remote_album_to_users.user_id AS user_id,
					remote_artists.tapesonic_artist_id AS artist_id,
					count(*) AS album_count
				FROM remote_albums
				JOIN remote_album_to_users ON remote_albums.id = remote_album_to_users.remote_album_id
				JOIN remote_artists ON remote_albums.remote_id = remote_artists.remote_id AND remote_albums.artist_id = remote_artists.artist_id
				GROUP BY remote_album_to_users.user_id, remote_artists.tapesonic_artist_id
			),
			artist_aggregate (user_id, artist_id, album_count) AS (
				SELECT
					user_id AS user_id,
					artist_id AS artist_id,
					sum(album_count) AS album_count
				FROM artist_raw_aggregate
				GROUP BY user_id, artist_id
			)

		SELECT
			artists.id AS id,
			artists.name AS name,
			NULL AS cover_id,
			artist_aggregate.album_count AS album_count,
			artists.search_name AS search_name,
			users.id AS user_id
		FROM artists
		JOIN users ON 1 = 1
		LEFT JOIN artist_aggregate ON users.id = artist_aggregate.user_id AND artists.id = artist_aggregate.artist_id
	`

	return store.db.Exec(sql).Error
}

func (store *ArtistStorage) FindArtistById(userId uuid.UUID, artistId string) (*model.LibraryArtist, error) {
	result := model.LibraryArtist{}

	sql := `
		SELECT
			all_artists.id AS id,
			all_artists.name AS name,
			all_artists.cover_id AS cover_id,
			all_artists.album_count AS album_count
		FROM all_artists
		WHERE all_artists.user_id = @userId AND all_artists.id = @artistId
		LIMIT 1
	`
	params := map[string]any{
		"userId":   userId,
		"artistId": artistId,
	}

	err := store.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}

func (store *ArtistStorage) SearchArtistsByQuery(userId uuid.UUID, query string, count int, offset int) ([]model.LibraryArtist, error) {
	filter := util.MakeTextSearchCondition([]string{"all_artists.search_name"}, query)
	if filter == "" {
		return []model.LibraryArtist{}, nil
	}

	sql := fmt.Sprintf(
		`
			SELECT
				all_artists.id AS id,
				all_artists.name AS name,
				all_artists.cover_id AS cover_id,
				all_artists.album_count AS album_count
			FROM all_artists
			WHERE all_artists.user_id = @userId AND %s
			ORDER BY all_artists.id
			LIMIT %d OFFSET %d
		`,
		filter,
		count,
		offset,
	)
	params := map[string]any{
		"userId": userId,
	}

	result := []model.LibraryArtist{}
	return result, store.db.Raw(sql, params).Find(&result).Error
}

func (store *ArtistStorage) GetArtistsSortId(userId uuid.UUID, count int, offset int) ([]model.LibraryArtist, error) {
	sql := fmt.Sprintf(
		`
			SELECT
				all_artists.id AS id,
				all_artists.name AS name,
				all_artists.cover_id AS cover_id,
				all_artists.album_count AS album_count
			FROM all_artists
			WHERE all_artists.user_id = @userId AND all_artists.album_count > 0
			ORDER BY all_artists.id ASC
			LIMIT %d OFFSET %d
		`,
		count,
		offset,
	)
	params := map[string]any{
		"userId": userId,
	}

	result := []model.LibraryArtist{}
	return result, store.db.Raw(sql, params).Find(&result).Error
}

func (store *ArtistStorage) GetArtistsSortName(userId uuid.UUID, count int, offset int) ([]model.LibraryArtist, error) {
	sql := fmt.Sprintf(
		`
			SELECT
				all_artists.id AS id,
				all_artists.name AS name,
				all_artists.cover_id AS cover_id,
				all_artists.album_count AS album_count
			FROM all_artists
			WHERE all_artists.user_id = @userId AND all_artists.album_count > 0
			ORDER BY lower(all_artists.name) ASC, all_artists.name ASC
			LIMIT %d OFFSET %d
		`,
		count,
		offset,
	)
	params := map[string]any{
		"userId": userId,
	}

	result := []model.LibraryArtist{}
	return result, store.db.Raw(sql, params).Find(&result).Error
}
