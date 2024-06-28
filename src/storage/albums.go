package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlbumStorage struct {
	db *gorm.DB
}

type Album struct {
	Id uuid.UUID

	Name        string
	Artist      string
	ReleaseDate *time.Time

	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	Tracks []*AlbumTrack
}

func (e *Album) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

type AlbumTrack struct {
	Id uuid.UUID

	AlbumId uuid.UUID
	Album   *Album

	TapeTrackId uuid.UUID
	TapeTrack   *TapeTrack

	TrackIndex int
}

func (e *AlbumTrack) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

func NewAlbumStorage(db *gorm.DB) (*AlbumStorage, error) {
	err := db.AutoMigrate(
		&Album{},
		&AlbumTrack{},
	)
	return &AlbumStorage{db: db}, err
}

func (storage *AlbumStorage) CreateAlbum(album *Album) error {
	for index, track := range album.Tracks {
		track.TrackIndex = index
	}

	return storage.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(album).Error
}

func (storage *AlbumStorage) UpdateAlbum(album *Album) (*Album, error) {
	return album, storage.db.Transaction(func(tx *gorm.DB) error {
		oldTrackIds := []uuid.UUID{}
		if err := tx.Model(&AlbumTrack{}).Where("album_id = ?", album.Id).Pluck("id", &oldTrackIds).Error; err != nil {
			return err
		}

		newTrackIds := []uuid.UUID{}
		for _, track := range album.Tracks {
			newTrackIds = append(newTrackIds, track.Id)
		}

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(album).Error; err != nil {
			return err
		}

		if len(newTrackIds) > 0 {
			if err := tx.Where("id IN ? AND id NOT IN ?", oldTrackIds, newTrackIds).Delete(&AlbumTrack{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("id IN ?", oldTrackIds).Delete(&AlbumTrack{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (storage *AlbumStorage) DeleteAlbum(id uuid.UUID) error {
	return storage.db.Select("Tracks").Delete(&Album{Id: id}).Error
}

func (storage *AlbumStorage) GetAllAlbums() ([]Album, error) {
	result := []Album{}
	// todo: get rid of preload
	return result, storage.db.Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Preload("Tracks.TapeTrack").Find(&result).Error
}

func (storage *AlbumStorage) GetSubsonicAlbum(id uuid.UUID) (*SubsonicAlbumItem, error) {
	albums, err := storage.getSubsonicAlbums(1, 0, fmt.Sprintf("albums.id = '%s'", id.String()), "albums.id")
	if err != nil {
		return nil, err
	}
	if len(albums) == 0 {
		return nil, fmt.Errorf("album with id %s doesn't exist", id.String())
	}

	return &albums[0], nil
}

func (storage *AlbumStorage) SearchSubsonicAlbums(count int, offset int, query []string) ([]SubsonicAlbumItem, error) {
	filter := []string{}
	for _, term := range query {
		searchField := "' ' || albums.artist || ' ' || albums.name"
		filter = append(filter, fmt.Sprintf("%s LIKE '%% %s%%' ESCAPE '%s'", searchField, EscapeTextLiteralForLike(term, "\\"), "\\"))
	}

	return storage.getSubsonicAlbums(count, offset, strings.Join(filter, " AND "), "albums.id")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortId(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "albums.id")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortRandom(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "random()")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortNewest(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "albums.created_at DESC")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortName(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "lower(albums.name)")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortArtist(count int, offset int) ([]SubsonicAlbumItem, error) {
	return storage.getSubsonicAlbums(count, offset, "", "lower(albums.artist)")
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
		order = "albums.release_date ASC"
	} else {
		fromYear, toYear = toYear, fromYear
		order = "albums.release_date DESC"
	}

	filter := "albums.release_date IS NOT NULL"
	filter += fmt.Sprintf(" AND cast(strftime('%%Y', albums.release_date) AS INTEGER) BETWEEN %d AND %d", fromYear, toYear)

	return storage.getSubsonicAlbums(count, offset, filter, order)
}

func (storage *AlbumStorage) getSubsonicAlbums(count int, offset int, filter string, order string) ([]SubsonicAlbumItem, error) {
	query := `
		WITH album_extra_info AS (
			SELECT
				album_tracks.album_id AS album_id,
				count(album_tracks.id) AS song_count,
				sum(tape_tracks.end_offset_ms - tape_tracks.start_offset_ms) / 1000 AS duration_sec,
				max(tape_track_listens.last_listened_at) AS last_listened_at,
				sum(tape_track_listens.listen_count) AS play_count,
				sum(tape_track_listens.listen_count * (tape_tracks.end_offset_ms - tape_tracks.start_offset_ms)) AS total_play_time
			FROM album_tracks
			LEFT JOIN tape_tracks ON tape_tracks.id = album_tracks.tape_track_id
			LEFT JOIN tape_track_listens ON tape_track_listens.tape_track_id = album_tracks.tape_track_id
			GROUP BY album_tracks.album_id
		)
		SELECT *
		FROM albums
		LEFT JOIN album_extra_info ON album_extra_info.album_id = albums.id
	`

	if filter != "" {
		query += fmt.Sprintf("\nWHERE %s", filter)
	}

	query += fmt.Sprintf("\nORDER BY %s", order)
	query += fmt.Sprintf("\nLIMIT %d OFFSET %d", count, offset)

	result := []SubsonicAlbumItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *AlbumStorage) GetAlbumWithoutTracks(id uuid.UUID) (*Album, error) {
	result := Album{}
	return &result, storage.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (storage *AlbumStorage) GetAlbumWithTracks(id uuid.UUID) (*Album, error) {
	result := Album{}

	return &result, storage.db.Where(&Tape{Id: id}).Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Preload("Tracks.TapeTrack").Take(&result).Error
}

func (storage *AlbumStorage) GetAlbumRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	tapeIdFilter := storage.db.Raw(
		"SELECT DISTINCT tapes.id "+
			"FROM tapes "+
			"JOIN tape_files ON tapes.id = tape_files.tape_id "+
			"JOIN tape_tracks ON tape_files.id = tape_tracks.tape_file_id "+
			"JOIN album_tracks ON tape_tracks.id = album_tracks.tape_track_id "+
			"WHERE album_tracks.album_id = ?",
		id,
	)

	err := storage.db.Model(&Tape{}).Where("id IN (?)", tapeIdFilter).Find(&result.Tapes).Error

	return &result, err
}
