package storage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TapeStorage struct {
	db *gorm.DB
}

type Tape struct {
	Id uuid.UUID

	Metadata string
	Url      string

	Name       string
	AuthorName string

	ThumbnailPath string

	Tracks []*TapeTrack
}

type TapeTrack struct {
	Id uuid.UUID

	TapeId uuid.UUID
	Tape   *Tape

	FilePath string

	RawStartOffsetMs int
	StartOffsetMs    int
	RawEndOffsetMs   int
	EndOffsetMs      int

	Artist string
	Title  string

	TrackIndex int
}

func NewTapeStorage(db *gorm.DB) (*TapeStorage, error) {
	err := db.AutoMigrate(
		&Tape{},
		&TapeTrack{},
	)
	return &TapeStorage{db: db}, err
}

func (storage *TapeStorage) UpsertTape(tape *Tape) error {
	for index, track := range tape.Tracks {
		track.TrackIndex = index
	}

	return storage.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tracks").Save(tape).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Model(tape).Association("Tracks").Unscoped().Replace(tape.Tracks); err != nil {
			return err
		}

		return nil
	})
}

func (storage *TapeStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	// todo: get rid of preload
	return result, storage.db.Preload("Tracks").Find(&result).Error
}

func (storage *TapeStorage) GetTapeWithoutTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}
	return &result, storage.db.Where(&Tape{Id: id}).Take(&result).Error
}

func (storage *TapeStorage) GetTapeWithTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}

	return &result, storage.db.Where(&Tape{Id: id}).Preload("Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Take(&result).Error
}

func (storage *TapeStorage) GetTapeRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	playlistIdFilter := storage.db.Raw(
		"SELECT DISTINCT playlist_id "+
			"FROM playlist_tracks "+
			"JOIN tape_tracks ON tape_tracks.id = playlist_tracks.tape_track_id "+
			"WHERE tape_tracks.tape_id = ?",
		id,
	)

	err := storage.db.Model(&Playlist{}).Where("id IN (?)", playlistIdFilter).Find(&result.Playlists).Error

	return &result, err
}

func (storage *TapeStorage) GetTapeTrack(id uuid.UUID) (*TapeTrack, error) {
	result := TapeTrack{}
	return &result, storage.db.Where(&TapeTrack{Id: id}).Take(&result).Error
}
