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

func (storage *TrackStorage) ReplaceTracksForSource(sourceId uuid.UUID, tracks []SourceTrack) ([]SourceTrack, error) {
	return tracks, storage.db.Transaction(func(tx *gorm.DB) error {
		upsertSql := `
			INSERT INTO source_tracks (id, source_id, start_offset_ms, end_offset_ms, artist, title, search_artist, search_title)
			VALUES (@id, @sourceId, @startOffsetMs, @endOffsetMs, @artist, @title, @searchArtist, @searchTitle)
			ON CONFLICT (id) DO UPDATE
			SET
				start_offset_ms = excluded.start_offset_ms,
				end_offset_ms = excluded.end_offset_ms,
				artist = excluded.artist,
				title = excluded.title,
				search_artist = excluded.search_artist,
				search_title = excluded.search_title
		`

		for _, track := range tracks {
			upsertParams := map[string]any{
				"id":            track.Id,
				"sourceId":      track.SourceId,
				"startOffsetMs": track.StartOffsetMs,
				"endOffsetMs":   track.EndOffsetMs,
				"artist":        track.Artist,
				"title":         track.Title,
				"searchArtist":  util.MakeTextSearchString(track.Artist),
				"searchTitle":   util.MakeTextSearchString(track.Title),
			}

			if err := tx.Exec(upsertSql, upsertParams).Error; err != nil {
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

		return nil
	})
}

func (storage *TrackStorage) GetDirectTracksBySource(sourceId uuid.UUID) ([]SourceTrack, error) {
	query := `
		SELECT *
		FROM source_tracks
		WHERE source_id = @id
		ORDER BY source_tracks.start_offset_ms ASC
	`
	params := map[string]any{
		"id": sourceId,
	}

	tracks := []SourceTrack{}
	return tracks, storage.db.Raw(query, params).Find(&tracks).Error
}

func (storage *TrackStorage) GetAllTracksBySource(sourceId uuid.UUID) ([]SourceTrack, error) {
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
			SELECT source_tracks.*
			FROM source_tracks
			JOIN all_sources ON all_sources.child_id = source_tracks.source_id
			ORDER BY all_sources.nest_level ASC, all_sources.list_index ASC, source_tracks.start_offset_ms ASC
		`,
		sourceId,
		sourceId,
	)

	tracks := []SourceTrack{}
	return tracks, storage.db.Raw(query).Find(&tracks).Error
}
