package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlbumStorage struct {
	db *gorm.DB
}

func NewAlbumStorage(db *gorm.DB) (*AlbumStorage, error) {
	err := db.AutoMigrate()
	return &AlbumStorage{db: db}, err
}

func (storage *AlbumStorage) GetSubsonicAlbum(id uuid.UUID) (*SubsonicAlbumItem, error) {
	albums, err := storage.getSubsonicAlbums(1, 0, fmt.Sprintf("tapes.id = '%s'", id.String()), "tapes.id")
	if err != nil {
		return nil, err
	}
	if len(albums) == 0 {
		return nil, fmt.Errorf("album with id %s doesn't exist", id.String())
	}

	return &albums[0], nil
}

func (storage *AlbumStorage) SearchSubsonicAlbums(count int, offset int, query string) ([]SubsonicAlbumItem, error) {
	filter := MakeTextSearchCondition([]string{"tapes.artist", "tapes.name"}, query)
	if filter == "" {
		return []SubsonicAlbumItem{}, nil
	}

	return storage.getSubsonicAlbums(count, offset, filter, "tapes.id")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortId(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "tapes.id")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortRandom(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "random()")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortNewest(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "tapes.created_at DESC")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortName(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "lower(tapes.name)")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortArtist(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "lower(tapes.artist)")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortRecent(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "album_extra_info.last_listened_at IS NOT NULL", "album_extra_info.last_listened_at DESC")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortFrequent(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "album_extra_info.total_play_time > 0", "album_extra_info.total_play_time DESC")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortReleaseDate(count int, offset int, fromYear int, toYear int) ([]SubsonicAlbumItem, error) {
	var order string
	if fromYear <= toYear {
		order = "tapes.released_at ASC"
	} else {
		fromYear, toYear = toYear, fromYear
		order = "tapes.released_at DESC"
	}

	filter := fmt.Sprintf("tapes.released_at IS NOT NULL AND cast(strftime('%%Y', tapes.released_at) AS INTEGER) BETWEEN %d AND %d", fromYear, toYear)

	return storage.getSubsonicAlbums(count, offset, filter, order)
}

func (storage *AlbumStorage) getSubsonicAlbums(count int, offset int, filter string, order string) ([]SubsonicAlbumItem, error) {
	query := `
		WITH album_extra_info AS (
			SELECT
				tape_to_tracks.tape_id AS tape_id,
				count(*) AS song_count,
				sum(tracks.end_offset_ms - tracks.start_offset_ms) / 1000 AS duration_sec,
				max(track_listens.last_listened_at) AS last_listened_at,
				sum(track_listens.listen_count) AS play_count,
				sum(track_listens.listen_count * (tracks.end_offset_ms - tracks.start_offset_ms)) AS total_play_time
			FROM tape_to_tracks
			JOIN tracks ON tracks.id = tape_to_tracks.track_id
			LEFT JOIN track_listens ON track_listens.track_id = tape_to_tracks.track_id
			GROUP BY tape_to_tracks.tape_id
		)
		SELECT
			tapes.id AS id,
			tapes.name AS name,
			tapes.artist AS artist,
			tapes.released_at AS release_date,
			tapes.thumbnail_id AS thumbnail_id,
			tapes.created_at AS created_at,
			tapes.updated_at AS updated_at,
			album_extra_info.song_count AS song_count,
			album_extra_info.duration_sec AS duration_sec,
			album_extra_info.play_count AS play_count
		FROM tapes
		LEFT JOIN album_extra_info ON album_extra_info.tape_id = tapes.id
	`

	conditions := []string{fmt.Sprintf("tapes.type = '%s'", TAPE_TYPE_ALBUM)}
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

	result := []SubsonicAlbumItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}
