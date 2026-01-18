package library

import (
	"errors"
	"fmt"
	"tapesonic/model"
	"tapesonic/storage"

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
			remote_artists_aggregate (user_id, id, album_count) AS (
				SELECT
					remote_album_to_users.user_id AS user_id,
					remote_artists.id AS id,
					count(*) AS album_count
				FROM remote_albums
				JOIN remote_album_to_users ON remote_albums.id = remote_album_to_users.remote_album_id
				JOIN remote_artists ON remote_albums.remote_id = remote_artists.remote_id AND remote_albums.artist_id = remote_artists.artist_id
				GROUP BY remote_album_to_users.user_id, remote_artists.id
			)

		SELECT
			remote_artists.id AS id,
			remote_artists.name AS name,
			remote_covers.id AS cover_id,
			remote_artists_aggregate.album_count AS album_count,
			remote_artists.search_name AS search_name,
			remote_artist_to_users.user_id AS user_id
		FROM remote_artists
		JOIN remote_artist_to_users ON remote_artists.id = remote_artist_to_users.remote_artist_id
		LEFT JOIN remote_artists_aggregate ON remote_artists.id = remote_artists_aggregate.id AND remote_artist_to_users.user_id = remote_artists_aggregate.user_id
		LEFT JOIN remote_covers ON remote_artists.remote_id = remote_covers.remote_id AND remote_artists.cover_id = remote_covers.cover_id
	`

	return store.db.Exec(sql).Error
}

func (store *ArtistStorage) FindArtistById(userId uuid.UUID, artistId string) (*model.LibraryArtist, error) {
	result := model.LibraryArtist{}

	params := map[string]any{
		"userId":   userId,
		"artistId": artistId,
	}

	err := store.db.Raw("SELECT * FROM all_artists WHERE user_id = @userId AND id = @artistId LIMIT 1", params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err == nil {
		return &result, nil
	} else {
		return nil, err
	}
}

func (store *ArtistStorage) SearchArtistsByQuery(userId uuid.UUID, query string, count int, offset int) ([]model.LibraryArtist, error) {
	filter := storage.MakeTextSearchCondition([]string{"all_artists.search_name"}, query)
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
			WHERE user_id = @userId AND %s
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
			WHERE user_id = @userId
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
