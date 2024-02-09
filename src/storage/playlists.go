package storage

import (
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
