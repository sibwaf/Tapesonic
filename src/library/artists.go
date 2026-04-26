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

func (store *ArtistStorage) FindArtistById(userId uuid.UUID, artistId string) (*model.LibraryArtist, error) {
	result := model.LibraryArtist{}

	sql := `
		SELECT
			all_artists.id AS id,
			all_artists.name AS name,
			all_artists.artwork_id AS artwork_id,
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
				all_artists.artwork_id AS artwork_id,
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
				all_artists.artwork_id AS artwork_id,
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
				all_artists.artwork_id AS artwork_id,
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
