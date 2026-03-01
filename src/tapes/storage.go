package tapes

import (
	"errors"
	"tapesonic/model"
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

func (store *TapeStorage) Create(tape Tape) (SavedTape, error) {
	result := SavedTape{}
	return result, store.db.Transaction(func(tx *gorm.DB) error {
		tapeSql := `
			INSERT INTO tapes (id, name, type, thumbnail_id, artist_id, released_at, created_at, updated_at, search_name)
			VALUES (@id, @name, @type, @thumbnailId, @artistId, @releasedAt, @createdAt, @updatedAt, @searchName)
			RETURNING *
		`
		tapeParams := map[string]any{
			"id":          tape.Id,
			"name":        tape.Name,
			"type":        tape.Type,
			"thumbnailId": tape.ThumbnailId,
			"artistId":    tape.ArtistId,
			"releasedAt":  tape.ReleasedAt,
			"createdAt":   tape.CreatedAt,
			"updatedAt":   tape.UpdatedAt,
			"searchName":  util.MakeTextSearchString(tape.Name),
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
				thumbnail_id = @thumbnailId,
				artist_id = @artistId,
				released_at = @releasedAt,
				updated_at = @updatedAt,
				search_name = @searchName
			WHERE id = @id
			RETURNING id
		`
		tapeParams := map[string]any{
			"id":          tape.Id,
			"name":        tape.Name,
			"type":        tape.Type,
			"thumbnailId": tape.ThumbnailId,
			"artistId":    tape.ArtistId,
			"releasedAt":  tape.ReleasedAt,
			"updatedAt":   tape.UpdatedAt,
			"searchName":  util.MakeTextSearchString(tape.Name),
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

func (store *TapeStorage) ReplaceTracksById(tapeId uuid.UUID, trackIds []uuid.UUID) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM tape_to_tracks WHERE tape_id = ?", tapeId).Error; err != nil {
			return err
		}

		sql := `
			INSERT INTO tape_to_tracks (tape_id, track_id, list_index)
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
			tapes.thumbnail_id AS thumbnail_id,
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
			tapes.thumbnail_id AS thumbnail_id,
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

func (store *TapeStorage) GetTracksById(tapeId uuid.UUID) ([]sources.SavedSourceTrack, error) {
	query := `
		SELECT
			source_tracks.id AS id,
			source_tracks.source_id AS source_id,
			source_tracks.start_offset_ms AS start_offset_ms,
			source_tracks.end_offset_ms AS end_offset_ms,
			artists.id AS artist_id,
			artists.name AS artist_name,
			source_tracks.title AS title
		FROM source_tracks
		JOIN tape_to_tracks ON tape_to_tracks.track_id = source_tracks.id
		LEFT JOIN artists ON source_tracks.artist_id = artists.id
		WHERE tape_to_tracks.tape_id = @tapeId
		ORDER BY tape_to_tracks.list_index
	`
	params := map[string]any{"tapeId": tapeId}

	tracks := []sources.SavedSourceTrack{}
	return tracks, store.db.Raw(query, params).Find(&tracks).Error
}

func (store *TapeStorage) GetTracksForMetadataGuessing(ids []uuid.UUID) ([]TrackForMetadataGuessing, error) {
	query := `
		SELECT
			source_tracks.id AS "id",
			sources.album_artist AS "album_artist",
			sources.album_title AS "album_title",
			sources.title AS "source_title",
			(
				SELECT json_group_array(parents.title)
				FROM sources parents
				JOIN source_hierarchies ON parents.id = source_hierarchies.parent_id
				WHERE source_hierarchies.child_id = sources.id
			) AS "source_parent_titles",
			source_tracks.artist_id AS "artist_id",
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
