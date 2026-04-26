package tapes

import (
	"errors"
	"tapesonic/model"
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

func (store *TapeStorage) Create(tape Tape) (SavedTape, error) {
	result := SavedTape{}
	return result, store.db.Transaction(func(tx *gorm.DB) error {
		tapeSql := `
			INSERT INTO tapes (id, name, type, artwork_id, artist_id, released_at, created_by, created_at, updated_at, search_name)
			VALUES (@id, @name, @type, @artworkId, @artistId, @releasedAt, @createdBy, @createdAt, @updatedAt, @searchName)
			RETURNING *
		`
		tapeParams := map[string]any{
			"id":         tape.Id,
			"name":       tape.Name,
			"type":       tape.Type,
			"artworkId":  tape.ArtworkId,
			"artistId":   tape.ArtistId,
			"releasedAt": tape.ReleasedAt,
			"createdBy":  tape.CreatedBy,
			"createdAt":  tape.CreatedAt,
			"updatedAt":  tape.UpdatedAt,
			"searchName": util.MakeTextSearchString(tape.Name),
		}

		if err := tx.Raw(tapeSql, tapeParams).Take(&tape).Error; err != nil {
			return err
		}

		innerResult, err := store.getByIdWithTx(tx, tape.Id)
		result = innerResult
		return err
	})
}

func (store *TapeStorage) Update(tape Tape) (SavedTape, error) {
	result := SavedTape{}
	return result, store.db.Transaction(func(tx *gorm.DB) error {
		tapeSql := `
			UPDATE tapes
			SET
				name = @name,
				type = @type,
				artwork_id = @artworkId,
				artist_id = @artistId,
				released_at = @releasedAt,
				updated_at = @updatedAt,
				search_name = @searchName
			WHERE id = @id
			RETURNING id
		`
		tapeParams := map[string]any{
			"id":         tape.Id,
			"name":       tape.Name,
			"type":       tape.Type,
			"artworkId":  tape.ArtworkId,
			"artistId":   tape.ArtistId,
			"releasedAt": tape.ReleasedAt,
			"updatedAt":  tape.UpdatedAt,
			"searchName": util.MakeTextSearchString(tape.Name),
		}

		err := tx.Raw(tapeSql, tapeParams).Take(&tape).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ErrNotFound
		} else if err != nil {
			return err
		}

		innerResult, err := store.getByIdWithTx(tx, tape.Id)
		result = innerResult
		return err
	})
}

func (store *TapeStorage) ReplaceTracksById(tapeId uuid.UUID, trackIds []string) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM tape_tracks WHERE tape_id = ?", tapeId).Error; err != nil {
			return err
		}

		sql := `
			INSERT INTO tape_tracks (tape_id, track_id, list_index)
			VALUES (@tapeId, @trackId, @listIndex)
		`

		for i, trackId := range trackIds {
			params := map[string]any{
				"tapeId":    tapeId,
				"trackId":   trackId,
				"listIndex": i,
			}

			if err := tx.Exec(sql, params).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (store *TapeStorage) DeleteById(id uuid.UUID) error {
	return store.db.Exec("DELETE FROM tapes WHERE id = @id", map[string]any{"id": id}).Error
}

func (store *TapeStorage) GetAllTapes() ([]SavedTape, error) {
	sql := `
		SELECT
			tapes.id AS id,
			tapes.type AS type,
			tapes.name AS name,
			tapes.artwork_id AS artwork_id,
			artists.id AS artist_id,
			artists.name AS artist_name,
			tapes.released_at AS released_at,
			tapes.created_at AS created_at
		FROM tapes
		LEFT JOIN artists ON tapes.artist_id = artists.id
		ORDER BY tapes.created_at DESC
	`

	result := []SavedTape{}
	return result, store.db.Raw(sql).Find(&result).Error
}

func (store *TapeStorage) GetById(id uuid.UUID) (SavedTape, error) {
	result := SavedTape{}
	return result, store.db.Transaction(func(tx *gorm.DB) error {
		innerResult, err := store.getByIdWithTx(tx, id)
		result = innerResult
		return err
	})
}

func (store *TapeStorage) getByIdWithTx(tx *gorm.DB, id uuid.UUID) (SavedTape, error) {
	sql := `
		SELECT
			tapes.id AS id,
			tapes.type AS type,
			tapes.name AS name,
			tapes.artwork_id AS artwork_id,
			artists.id AS artist_id,
			artists.name AS artist_name,
			tapes.released_at AS released_at,
			tapes.created_at AS created_at
		FROM tapes
		LEFT JOIN artists ON tapes.artist_id = artists.id
		WHERE tapes.id = @id
	`
	params := map[string]any{"id": id}

	result := SavedTape{}
	err := tx.Raw(sql, params).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return SavedTape{}, model.ErrNotFound
	} else {
		return result, err
	}
}

func (store *TapeStorage) GetTrackIdsById(tapeId uuid.UUID) ([]string, error) {
	query := `
		SELECT tape_tracks.track_id
		FROM tape_tracks
		WHERE tape_tracks.tape_id = @tapeId
		ORDER BY tape_tracks.list_index
	`
	params := map[string]any{"tapeId": tapeId}

	tracks := []string{}
	return tracks, store.db.Raw(query, params).Find(&tracks).Error
}
