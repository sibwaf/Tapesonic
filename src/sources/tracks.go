package sources

import (
	"errors"
	"fmt"
	"tapesonic/model"
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
					source_hierarchy.parent_id,
					source_hierarchy.child_id,
					all_sources.nest_level + 1,
					source_hierarchy.list_index
				FROM source_hierarchy
				JOIN all_sources ON all_sources.child_id = source_hierarchy.parent_id
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

func (storage *TrackStorage) GetSourceDescriptor(trackId uuid.UUID) (SourceTrackFileDescriptor, error) {
	query := `
		SELECT
			source_files.media_path AS local_path,
			source_files.format AS local_format,
			source_files.codec AS local_codec,
			sources.url AS remote_url,
			sources.duration_ms AS source_duration_ms,
			source_tracks.start_offset_ms AS start_offset_ms,
			source_tracks.end_offset_ms AS end_offset_ms
		FROM source_tracks
		JOIN sources ON sources.id = source_tracks.source_id
		LEFT JOIN source_files ON source_files.source_id = sources.id
		WHERE source_tracks.id = @trackId
	`
	params := map[string]any{"trackId": trackId}

	sourceDescriptor := SourceTrackFileDescriptor{}
	err := storage.db.Raw(query, params).Take(&sourceDescriptor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sourceDescriptor, model.ErrNotFound
	} else {
		return sourceDescriptor, err
	}
}

func (storage *TrackStorage) FindTracksForMetadataGuessingByIds(trackIds []string) ([]SourceTrackForMetadataGuessing, error) {
	if len(trackIds) == 0 {
		return []SourceTrackForMetadataGuessing{}, nil
	}

	sql := `
		SELECT
			source_tracks.id AS "id",
			sources.album_artist AS "album_artist",
			sources.album_title AS "album_title",
			sources.title AS "source_title",
			(
				SELECT json_group_array(parents.title)
				FROM sources parents
				JOIN source_hierarchy ON parents.id = source_hierarchy.parent_id
				WHERE source_hierarchy.child_id = sources.id
			) AS "source_parent_titles",
			source_tracks.artist_id AS "artist_id",
			sources.release_date AS "release_date",
			sources.artwork_id AS "artwork_id"
		FROM source_tracks
		JOIN sources ON source_tracks.source_id = sources.id
		WHERE source_tracks.id IN @trackIds
	`
	params := map[string]any{
		"trackIds": trackIds,
	}

	result := []SourceTrackForMetadataGuessing{}
	return result, storage.db.Raw(sql, params).Find(&result).Error
}
