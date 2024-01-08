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

	Files []*TapeFile
}

func (e *Tape) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

type TapeFile struct {
	Id uuid.UUID

	TapeId uuid.UUID
	Tape   *Tape

	Metadata string
	Url      string

	Name       string
	AuthorName string

	ThumbnailPath string
	MediaPath     string

	FileIndex int

	Tracks []*TapeTrack
}

func (e *TapeFile) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

type TapeTrack struct {
	Id uuid.UUID

	TapeFileId uuid.UUID
	TapeFile   *TapeFile

	RawStartOffsetMs int
	StartOffsetMs    int
	RawEndOffsetMs   int
	EndOffsetMs      int

	Artist string
	Title  string

	TrackIndex int
}

func (e *TapeTrack) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

func NewTapeStorage(db *gorm.DB) (*TapeStorage, error) {
	err := db.AutoMigrate(
		&Tape{},
		&TapeFile{},
		&TapeTrack{},
	)

	return &TapeStorage{db: db}, err
}

func (storage *TapeStorage) UpsertTape(tape *Tape) error {
	newFileIds := []uuid.UUID{}
	newTrackIds := []uuid.UUID{}

	for fileIndex, file := range tape.Files {
		file.FileIndex = fileIndex
		newFileIds = append(newFileIds, file.Id)

		for trackIndex, track := range file.Tracks {
			track.TrackIndex = trackIndex
			newTrackIds = append(newTrackIds, track.Id)
		}
	}

	return storage.db.Transaction(func(tx *gorm.DB) error {
		// prepare

		oldFileIds := []uuid.UUID{}
		if err := tx.Model(&TapeFile{}).Where("tape_id = ?", tape.Id).Pluck("id", &oldFileIds).Error; err != nil {
			return err
		}

		oldTrackIds := []uuid.UUID{}
		if err := tx.Model(&TapeTrack{}).Where("tape_file_id IN ?", oldFileIds).Pluck("id", &oldTrackIds).Error; err != nil {
			return err
		}

		// save

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(tape).Error; err != nil {
			return err
		}

		// cleanup

		if len(newTrackIds) > 0 {
			if err := tx.Where("id IN ? AND id NOT IN ?", oldTrackIds, newTrackIds).Delete(&TapeTrack{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("id IN ?", oldTrackIds).Delete(&TapeTrack{}).Error; err != nil {
				return err
			}
		}

		if len(newFileIds) > 0 {
			if err := tx.Where("id IN ? AND id NOT IN ?", oldFileIds, newFileIds).Delete(&TapeFile{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Where("id IN ?", oldFileIds).Delete(&TapeFile{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (storage *TapeStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	return result, storage.db.Find(&result).Error
}

func (storage *TapeStorage) GetTapeWithFiles(id uuid.UUID) (*Tape, error) {
	result := Tape{}

	return &result, storage.db.Where(&Tape{Id: id}).Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("file_index ASC")
	}).Take(&result).Error
}

func (storage *TapeStorage) GetTapeWithFilesAndTracks(id uuid.UUID) (*Tape, error) {
	result := Tape{}

	return &result, storage.db.Where(&Tape{Id: id}).Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Order("file_index ASC")
	}).Preload("Files.Tracks", func(db *gorm.DB) *gorm.DB {
		return db.Order("track_index ASC")
	}).Take(&result).Error
}

func (storage *TapeStorage) GetTapeRelationships(id uuid.UUID) (*RelatedItems, error) {
	result := RelatedItems{}

	playlistIdFilter := storage.db.Raw(
		"SELECT DISTINCT playlists.id "+
			"FROM playlists "+
			"JOIN playlist_tracks ON playlist_tracks.playlist_id = playlists.id "+
			"JOIN tape_tracks ON tape_tracks.id = playlist_tracks.tape_track_id "+
			"JOIN tape_files ON tape_files.id = tape_tracks.tape_file_id "+
			"WHERE tape_files.tape_id = ?",
		id,
	)

	err := storage.db.Model(&Playlist{}).Where("id IN (?)", playlistIdFilter).Find(&result.Playlists).Error

	return &result, err
}

func (storage *TapeStorage) GetTapeTrackWithFile(id uuid.UUID) (*TapeTrack, error) {
	result := TapeTrack{}
	return &result, storage.db.Where(&TapeTrack{Id: id}).Preload("TapeFile").Take(&result).Error
}
