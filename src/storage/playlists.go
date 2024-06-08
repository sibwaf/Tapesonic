package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaylistStorage struct {
	db *gorm.DB
}

type Playlist struct {
	Id uuid.UUID

	Name          string
	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	Tracks []*PlaylistTrack
}

func (e *Playlist) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

type PlaylistTrack struct {
	Id uuid.UUID

	PlaylistId uuid.UUID
	Playlist   *Playlist

	TapeTrackId uuid.UUID
	TapeTrack   *TapeTrack

	TrackIndex int
}

func (e *PlaylistTrack) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

func NewPlaylistStorage(db *gorm.DB) (*PlaylistStorage, error) {
	err := db.AutoMigrate(
		&Playlist{},
		&PlaylistTrack{},
	)
	return &PlaylistStorage{db: db}, err
}

func (storage *PlaylistStorage) CreatePlaylist(playlist *Playlist) error {
	for index, track := range playlist.Tracks {
		track.TrackIndex = index
	}

	return storage.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(playlist).Error
}

func (storage *PlaylistStorage) DeletePlaylist(id uuid.UUID) error {
	return storage.db.Select("Tracks").Delete(&Playlist{Id: id}).Error
}

func (storage *PlaylistStorage) GetAllPlaylists() ([]Playlist, error) {
	result := []Playlist{}
	// todo: get rid of preload
	return result, storage.db.Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Preload("Tracks.TapeTrack").Find(&result).Error
}

func (storage *PlaylistStorage) GetPlaylistWithoutTracks(id uuid.UUID) (*Playlist, error) {
	result := Playlist{}
	return &result, storage.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (storage *PlaylistStorage) GetPlaylistWithTracks(id uuid.UUID) (*Playlist, error) {
	result := Playlist{}

	return &result, storage.db.Where(&Tape{Id: id}).Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Preload("Tracks.TapeTrack").Take(&result).Error
}

func (storage *PlaylistStorage) GetSubsonicPlaylist(id uuid.UUID) (*SubsonicPlaylistItem, error) {
	playlists, err := storage.getSubsonicPlaylists(1, 0, fmt.Sprintf("playlists.id = '%s'", id.String()), "playlists.id")
	if err != nil {
		return nil, err
	}
	if len(playlists) == 0 {
		return nil, fmt.Errorf("playlist with id %s doesn't exist", id.String())
	}

	return &playlists[0], nil
}

func (storage *PlaylistStorage) GetSubsonicPlaylists(count int, offset int) ([]SubsonicPlaylistItem, error) {
	return storage.getSubsonicPlaylists(count, offset, "", "playlists.updated_at DESC")
}

func (storage *PlaylistStorage) getSubsonicPlaylists(count int, offset int, filter string, order string) ([]SubsonicPlaylistItem, error) {
	query := `
		WITH playlist_extra_info AS (
			SELECT
				playlist_tracks.playlist_id AS playlist_id,
				count(playlist_tracks.id) AS song_count,
				sum(tape_tracks.end_offset_ms - tape_tracks.start_offset_ms) / 1000 AS duration_sec
			FROM playlist_tracks
			LEFT JOIN tape_tracks ON tape_tracks.id = playlist_tracks.tape_track_id
			LEFT JOIN tape_track_listens ON tape_track_listens.tape_track_id = playlist_tracks.tape_track_id
			GROUP BY playlist_tracks.playlist_id
		)
		SELECT *
		FROM playlists
		LEFT JOIN playlist_extra_info ON playlist_extra_info.playlist_id = playlists.id
	`

	if filter != "" {
		query += fmt.Sprintf(" WHERE %s", filter)
	}

	query += fmt.Sprintf(" ORDER BY %s", order)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", count, offset)

	result := []SubsonicPlaylistItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *PlaylistStorage) GetPlaylistRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	tapeIdFilter := storage.db.Raw(
		"SELECT DISTINCT tapes.id "+
			"FROM tapes "+
			"JOIN tape_files ON tapes.id = tape_files.tape_id "+
			"JOIN tape_tracks ON tape_files.id = tape_tracks.tape_file_id "+
			"JOIN playlist_tracks ON tape_tracks.id = playlist_tracks.tape_track_id "+
			"WHERE playlist_tracks.playlist_id = ?",
		id,
	)

	err := storage.db.Model(&Tape{}).Where("id IN (?)", tapeIdFilter).Find(&result.Tapes).Error

	return &result, err
}
