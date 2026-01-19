package storage

import (
	"fmt"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Track struct {
	Id uuid.UUID

	SourceId uuid.UUID
	Source   *Source

	StartOffsetMs int64
	EndOffsetMs   int64

	Artist string
	Title  string

	SearchArtist string
	SearchTitle  string
}

func (e *Track) BeforeCreate(tx *gorm.DB) error {
	if e.Id.ID() == 0 {
		e.Id = uuid.New()
	}
	return nil
}

type TrackStorage struct {
	db *DbHelper
}

func NewTrackStorage(db *gorm.DB) (*TrackStorage, error) {
	if err := db.AutoMigrate(&Track{}); err != nil {
		return nil, err
	}

	return &TrackStorage{db: NewDbHelper(db)}, nil
}

func (storage *TrackStorage) ReplaceTracksForSource(sourceId uuid.UUID, tracks []Track) ([]Track, error) {
	return tracks, storage.db.Transaction(func(tx *gorm.DB) error {
		for i := range tracks {
			tracks[i].SourceId = sourceId
			tracks[i].Source = nil
			tracks[i].SearchArtist = util.MakeTextSearchString(tracks[i].Artist)
			tracks[i].SearchTitle = util.MakeTextSearchString(tracks[i].Title)
		}

		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}, clause.Returning{}).Save(&tracks).Error; err != nil {
			return err
		}

		trackIds := []uuid.UUID{}
		for _, track := range tracks {
			trackIds = append(trackIds, track.Id)
		}

		if err := tx.Where("source_id = ? AND id NOT IN ?", sourceId.String(), trackIds).Delete(&Track{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (storage *TrackStorage) GetDirectTracksBySource(sourceId uuid.UUID) ([]Track, error) {
	tracks := []Track{}
	return tracks, storage.db.Order("tracks.start_offset_ms ASC").Find(&tracks, fmt.Sprintf("tracks.source_id = '%s'", sourceId.String())).Error
}

func (storage *TrackStorage) GetAllTracksBySource(sourceId uuid.UUID) ([]Track, error) {
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
			SELECT tracks.*
			FROM tracks
			JOIN all_sources ON all_sources.child_id = tracks.source_id
			ORDER BY all_sources.nest_level ASC, all_sources.list_index ASC, tracks.start_offset_ms ASC
		`,
		sourceId,
		sourceId,
	)

	tracks := []Track{}
	return tracks, storage.db.Raw(query).Find(&tracks).Error
}

func (storage *TrackStorage) GetTracksByTape(tapeId uuid.UUID) ([]Track, error) {
	query := `
		SELECT tracks.*
		FROM tracks
		JOIN tape_to_tracks ON tape_to_tracks.track_id = tracks.id
		WHERE tape_to_tracks.tape_id = ?
		ORDER BY tape_to_tracks.list_index
	`

	tracks := []Track{}
	return tracks, storage.db.Raw(query, tapeId).Find(&tracks).Error
}

func (storage *TrackStorage) GetTracksForTapeMetadataGuessing(ids []uuid.UUID) ([]TrackForTapeMetadataGuessing, error) {
	query := `
		SELECT
			tracks.id AS "id",
			tracks.artist AS "artist",
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
		FROM tracks
		JOIN sources ON tracks.source_id = sources.id
		WHERE tracks.id IN @ids
	`
	params := map[string]any{
		"ids": ids,
	}

	tracks := []TrackForTapeMetadataGuessing{}
	return tracks, storage.db.Raw(query, params).Find(&tracks).Error
}

func (storage *TrackStorage) GetTracksWithSourcesByIds(ids []uuid.UUID) ([]Track, error) {
	tracks := []Track{}
	return tracks, storage.db.Preload("Source").Where("id IN ?", ids).Find(&tracks).Error
}
