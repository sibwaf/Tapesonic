package tapes

import (
	"tapesonic/sources"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TapeStorage struct {
	db *gorm.DB
}

func newTapeStorage(db *gorm.DB) *TapeStorage {
	return &TapeStorage{db: db}
}

func (store *TapeStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&Tape{}, &TapeToTrack{})
}

func (store *TapeStorage) Create(tape Tape) (Tape, error) {
	for i := range tape.Tracks {
		tape.Tracks[i].ListIndex = i
		tape.Tracks[i].Track = nil
		tape.Tracks[i].Tape = nil
		tape.Tracks[i].TapeId = tape.Id
	}

	err := store.db.Transaction(func(tx *gorm.DB) error {
		tapeSql := `
			INSERT INTO tapes (id, name, type, thumbnail_id, artist, released_at, created_at, updated_at, search_name)
			VALUES (@id, @name, @type, @thumbnailId, @artist, @releasedAt, @createdAt, @updatedAt, @searchName)
			RETURNING *
		`
		tapeParams := map[string]any{
			"id":          tape.Id,
			"name":        tape.Name,
			"type":        tape.Type,
			"thumbnailId": tape.ThumbnailId,
			"artist":      tape.Artist,
			"releasedAt":  tape.ReleasedAt,
			"createdAt":   tape.CreatedAt,
			"updatedAt":   tape.UpdatedAt,
			"searchName":  util.MakeTextSearchString(tape.Name),
		}

		if err := tx.Raw(tapeSql, tapeParams).First(&tape).Error; err != nil {
			return err
		}
		if len(tape.Tracks) > 0 {
			if err := tx.Save(&tape.Tracks).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return tape, err
}

func (store *TapeStorage) Update(tape Tape) (Tape, error) {
	for i := range tape.Tracks {
		tape.Tracks[i].ListIndex = i
		tape.Tracks[i].Track = nil
		tape.Tracks[i].Tape = nil
		tape.Tracks[i].TapeId = tape.Id
	}

	err := store.db.Transaction(func(tx *gorm.DB) error {
		tapeSql := `
			UPDATE tapes
			SET
				name = @name,
				type = @type,
				thumbnail_id = @thumbnailId,
				artist = @artist,
				released_at = @releasedAt,
				updated_at = @updatedAt,
				search_name = @searchName
			WHERE id = @id
			RETURNING *
		`
		tapeParams := map[string]any{
			"id":          tape.Id,
			"name":        tape.Name,
			"type":        tape.Type,
			"thumbnailId": tape.ThumbnailId,
			"artist":      tape.Artist,
			"releasedAt":  tape.ReleasedAt,
			"updatedAt":   tape.UpdatedAt,
			"searchName":  util.MakeTextSearchString(tape.Name),
		}

		if err := tx.Raw(tapeSql, tapeParams).Find(&tape).Error; err != nil {
			return err
		}
		if err := tx.Where("tape_id = ?", tape.Id).Delete(&TapeToTrack{}).Error; err != nil {
			return err
		}
		if len(tape.Tracks) > 0 {
			if err := tx.Save(&tape.Tracks).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return tape, err
}

func (store *TapeStorage) DeleteById(id uuid.UUID) error {
	return store.db.Exec("DELETE FROM tapes WHERE id = @id", map[string]any{"id": id}).Error
}

func (store *TapeStorage) GetAllTapes() ([]Tape, error) {
	result := []Tape{}
	return result, store.db.Order("created_at DESC").Find(&result).Error
}

func (store *TapeStorage) GetTape(id uuid.UUID) (Tape, error) {
	result := Tape{Id: id}
	return result, store.db.Find(&result).Error
}

func (store *TapeStorage) GetTracksById(tapeId uuid.UUID) ([]sources.SourceTrack, error) {
	query := `
		SELECT source_tracks.*
		FROM source_tracks
		JOIN tape_to_tracks ON tape_to_tracks.track_id = source_tracks.id
		WHERE tape_to_tracks.tape_id = ?
		ORDER BY tape_to_tracks.list_index
	`

	tracks := []sources.SourceTrack{}
	return tracks, store.db.Raw(query, tapeId).Find(&tracks).Error
}

func (store *TapeStorage) GetTracksForMetadataGuessing(ids []uuid.UUID) ([]TrackForMetadataGuessing, error) {
	query := `
		SELECT
			source_tracks.id AS "id",
			source_tracks.artist AS "artist",
			sources.title AS "source_title",
			(
				SELECT json_group_array(parents.title)
				FROM sources parents
				JOIN source_hierarchies ON parents.id = source_hierarchies.parent_id
				WHERE source_hierarchies.child_id = sources.id
			) AS "source_parent_titles",
			sources.album_artist AS "album_artist",
			sources.album_title AS "album_title",
			sources.release_date AS "release_date",
			sources.thumbnail_id AS "thumbnail_id"
		FROM source_tracks
		JOIN sources ON source_tracks.source_id = sources.id
		WHERE source_tracks.id IN @ids
	`
	params := map[string]any{
		"ids": ids,
	}

	tracks := []TrackForMetadataGuessing{}
	return tracks, store.db.Raw(query, params).Find(&tracks).Error
}
