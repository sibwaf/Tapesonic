package sources

import (
	"fmt"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackStorage struct {
	db *gorm.DB
}

func newTrackStorage(db *gorm.DB) *TrackStorage {
	return &TrackStorage{db: db}
}

func (storage *TrackStorage) PrepareDatabase() error {
	return storage.db.AutoMigrate(&SourceTrack{})
}

func (storage *TrackStorage) ReplaceTracksForSource(sourceId uuid.UUID, tracks []SourceTrack) ([]SavedSourceTrack, error) {
	result := []SavedSourceTrack{}
	return result, storage.db.Transaction(func(tx *gorm.DB) error {
		upsertSql := `
			INSERT INTO source_tracks (id, source_id, start_offset_ms, end_offset_ms, artist_id, title, search_title)
			VALUES (@id, @sourceId, @startOffsetMs, @endOffsetMs, @artistId, @title, @searchTitle)
			ON CONFLICT (id) DO UPDATE
			SET
				source_id = excluded.source_id,
				start_offset_ms = excluded.start_offset_ms,
				end_offset_ms = excluded.end_offset_ms,
				artist_id = excluded.artist_id,
				title = excluded.title,
				search_title = excluded.search_title
			RETURNING *
		`
		idUpsertSql := `
			INSERT INTO all_track_ids (id, source_track_id)
			VALUES (@id, @id)
			ON CONFLICT (id) DO NOTHING
		`

		for i, track := range tracks {
			upsertParams := map[string]any{
				"id":            track.Id,
				"sourceId":      sourceId,
				"startOffsetMs": track.StartOffsetMs,
				"endOffsetMs":   track.EndOffsetMs,
				"artistId":      track.ArtistId,
				"title":         track.Title,
				"searchTitle":   util.MakeTextSearchString(track.Title),
			}

			if err := tx.Raw(upsertSql, upsertParams).Take(&tracks[i]).Error; err != nil {
				return err
			}

			if err := tx.Exec(idUpsertSql, map[string]any{"id": tracks[i].Id}).Error; err != nil {
				return err
			}
		}

		trackIds := []uuid.UUID{}
		for _, track := range tracks {
			trackIds = append(trackIds, track.Id)
		}

		// todo: delete by sync tag, not ids
		if err := tx.Where("source_id = ? AND id NOT IN ?", sourceId.String(), trackIds).Delete(&SourceTrack{}).Error; err != nil {
			return err
		}

		innerResult, err := storage.getDirectTracksBySourceWithTx(tx, sourceId)
		result = innerResult
		return err
	})
}

func (storage *TrackStorage) GetDirectTracksBySource(sourceId uuid.UUID) ([]SavedSourceTrack, error) {
	result := []SavedSourceTrack{}
	return result, storage.db.Transaction(func(tx *gorm.DB) error {
		innerResult, err := storage.getDirectTracksBySourceWithTx(tx, sourceId)
		result = innerResult
		return err
	})
}

func (storage *TrackStorage) getDirectTracksBySourceWithTx(tx *gorm.DB, sourceId uuid.UUID) ([]SavedSourceTrack, error) {
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
		LEFT JOIN artists ON source_tracks.artist_id = artists.id
		WHERE source_id = @id
		ORDER BY source_tracks.start_offset_ms ASC
	`
	params := map[string]any{
		"id": sourceId,
	}

	tracks := []SavedSourceTrack{}
	return tracks, tx.Raw(query, params).Find(&tracks).Error
}

func (storage *TrackStorage) GetAllTracksBySource(sourceId uuid.UUID) ([]SavedSourceTrack, error) {
	query := fmt.Sprintf(
		`
			WITH RECURSIVE all_sources (parent_id, child_id, nest_level, list_index) AS (
				VALUES ('%s', '%s', 0, 0)
				UNION
				SELECT
					source_hierarchies.parent_id,
					source_hierarchies.child_id,
					all_sources.nest_level + 1,
					source_hierarchies.list_index
				FROM source_hierarchies
				JOIN all_sources ON all_sources.child_id = source_hierarchies.parent_id
			)
			SELECT
				source_tracks.id AS id,
				source_tracks.source_id AS source_id,
				source_tracks.start_offset_ms AS start_offset_ms,
				source_tracks.end_offset_ms AS end_offset_ms,
				artists.id AS artist_id,
				artists.name AS artist_name,
				source_tracks.title AS title
			FROM source_tracks
			JOIN all_sources ON all_sources.child_id = source_tracks.source_id
			LEFT JOIN artists ON source_tracks.artist_id = artists.id
			ORDER BY all_sources.nest_level ASC, all_sources.list_index ASC, source_tracks.start_offset_ms ASC
		`,
		sourceId,
		sourceId,
	)

	tracks := []SavedSourceTrack{}
	return tracks, storage.db.Raw(query).Find(&tracks).Error
}
