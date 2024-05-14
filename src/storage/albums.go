package storage

import (
	"fmt"
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

func (storage *AlbumStorage) GetSubsonicAlbumsSortRandom(count int, offset int) ([]SubsonicAlbumListItem, error) {
	return storage.getSubsonicAlbums(count, offset, "random()")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortNewest(count int, offset int) ([]SubsonicAlbumListItem, error) {
	return storage.getSubsonicAlbums(count, offset, "albums.created_at DESC")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortName(count int, offset int) ([]SubsonicAlbumListItem, error) {
	return storage.getSubsonicAlbums(count, offset, "lower(albums.name)")
}

func (storage *AlbumStorage) GetSubsonicAlbumsSortArtist(count int, offset int) ([]SubsonicAlbumListItem, error) {
	return storage.getSubsonicAlbums(count, offset, "lower(albums.artist)")
}

func (storage *AlbumStorage) getSubsonicAlbums(count int, offset int, order string) ([]SubsonicAlbumListItem, error) {
	query := `
		WITH album_extra_info AS (
			SELECT
				album_tracks.album_id AS album_id,
				count(album_tracks.id) AS song_count,
				sum(tape_tracks.end_offset_ms - tape_tracks.start_offset_ms) / 1000 AS duration_sec
			FROM album_tracks
			LEFT JOIN tape_tracks ON tape_tracks.id = album_tracks.tape_track_id
			GROUP BY album_tracks.album_id
		)
		SELECT *
		FROM albums
		LEFT JOIN album_extra_info ON album_extra_info.album_id = albums.id
	`
	query += fmt.Sprintf(" ORDER BY %s", order)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", count, offset)

	result := []SubsonicAlbumListItem{}
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
