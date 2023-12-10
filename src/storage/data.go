package storage

import (
	"os"
	"path"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DataStorage struct {
	db *gorm.DB
}

func NewDataStorage(
	dir string,
) (*DataStorage, error) {
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(path.Join(dir, "data.sqlite")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(
		&Tape{},
		&TapeTrack{},
		&Playlist{},
		&PlaylistTrack{},
	); err != nil {
		return nil, err
	}

	return &DataStorage{
		db: db,
	}, nil
}

func (ds *DataStorage) UpsertTape(tape *Tape) error {
	for index, track := range tape.Tracks {
		track.TrackIndex = index
	}

	return ds.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tracks").Save(tape).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(tape).Association("Tracks").Unscoped().Replace(tape.Tracks); err != nil {
			return err
		}

		return nil
	})
}

func (ds *DataStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	// todo: get rid of preload
	return result, ds.db.Preload("Tracks").Find(&result).Error
}

func (ds *DataStorage) GetTapeWithoutTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}
	return &result, ds.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (ds *DataStorage) GetTapeWithTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}

	return &result, ds.db.Where(&Tape{Id: id}).Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Take(&result).Error
}

func (ds *DataStorage) GetTapeRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	playlistIdFilter := ds.db.Raw(
		"SELECT DISTINCT playlist_id "+
			"FROM playlist_tracks "+
			"JOIN tape_tracks ON tape_tracks.id = playlist_tracks.tape_track_id "+
			"WHERE tape_tracks.tape_id = ?",
		id,
	)

	err := ds.db.Model(&Playlist{}).Where("id IN (?)", playlistIdFilter).Find(&result.Playlists).Error

	return &result, err
}

func (ds *DataStorage) GetTapeTrack(id uuid.UUID) (*TapeTrack, error) {
	result := TapeTrack{}
	return &result, ds.db.Where(&TapeTrack{Id: id}).Take(&result).Error
}

func (ds *DataStorage) CreatePlaylist(playlist *Playlist) error {
	for index, track := range playlist.Tracks {
		track.TrackIndex = index
	}

	return ds.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(playlist).Error
}

func (ds *DataStorage) DeletePlaylist(id uuid.UUID) error {
	return ds.db.Delete(&Playlist{}, id).Error
}

func (ds *DataStorage) GetAllPlaylists() ([]Playlist, error) {
	result := []Playlist{}
	// todo: get rid of preload
	return result, ds.db.Preload("Tracks").Preload("Tracks.TapeTrack").Find(&result).Error
}

func (ds *DataStorage) GetPlaylistWithoutTracks(id uuid.UUID) (*Playlist, error) {
	result := Playlist{}
	return &result, ds.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (ds *DataStorage) GetPlaylistWithTracks(id uuid.UUID) (*Playlist, error) {
	result := Playlist{}

	return &result, ds.db.Where(&Tape{Id: id}).Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Preload("Tracks.TapeTrack").Take(&result).Error
}

func (ds *DataStorage) GetPlaylistRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	tapeIdFilter := ds.db.Raw(
		"SELECT DISTINCT tape_id "+
			"FROM tape_tracks "+
			"JOIN playlist_tracks ON tape_tracks.id = playlist_tracks.tape_track_id "+
			"WHERE playlist_tracks.playlist_id = ?",
		id,
	)

	err := ds.db.Model(&Tape{}).Where("id IN (?)", tapeIdFilter).Find(&result.Tapes).Error

	return &result, err
}
