package storage

import (
	"fmt"
	"strings"

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

func (storage *TrackStorage) GetSubsonicTrack(id uuid.UUID) (*SubsonicTrackItem, error) {
	tracks, err := storage.getSubsonicTracks(1, 0, fmt.Sprintf("id = '%s'", id.String()), "id")
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, fmt.Errorf("track with id %s doesn't exist", id.String())
	}

	return &tracks[0], nil
}

func (storage *TrackStorage) SearchSubsonicTracks(count int, offset int, query string) ([]SubsonicTrackItem, error) {
	filter := MakeTextSearchCondition([]string{"artist", "album", "title"}, query)
	if filter == "" {
		return []SubsonicTrackItem{}, nil
	}

	return storage.getSubsonicTracks(count, offset, filter, "id")
}

func (storage *TrackStorage) GetSubsonicTracksSortId(count int, offset int) ([]SubsonicTrackItem, error) {
	return storage.getSubsonicTracks(count, offset, "", "id")
}

func (storage *TrackStorage) GetSubsonicTracksSortRandom(count int, fromYear *int, toYear *int) ([]SubsonicTrackItem, error) {
	conditions := []string{}

	if fromYear != nil && toYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', album_release_date) AS INTEGER) BETWEEN %d AND %d", *fromYear, *toYear))
	} else if fromYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', album_release_date) AS INTEGER) >= %d", *fromYear))
	} else if toYear != nil {
		conditions = append(conditions, fmt.Sprintf("cast(strftime('%%Y', album_release_date) AS INTEGER) <= %d", *toYear))
	}

	return storage.getSubsonicTracks(count, 0, strings.Join(conditions, " AND "), "random()")
}

func (storage *TrackStorage) GetSubsonicTracksByAlbum(albumId uuid.UUID) ([]SubsonicTrackItem, error) {
	query := fmt.Sprintf(
		`
			SELECT
				tracks.id AS id,
				sources.thumbnail_id AS thumbnail_id,
				tapes.id AS album_id,
				tapes.thumbnail_id AS album_thumbnail_id,
				tape_to_tracks.list_index AS album_track_index,
				tape_to_tracks.list_index AS playlist_track_index,
				tapes.name AS album,
				tracks.artist AS artist,
				tracks.title AS title,
				(tracks.end_offset_ms - tracks.start_offset_ms) / 1000 AS duration_sec,
				track_listens.listen_count AS play_count
			FROM tracks
			JOIN sources ON sources.id = tracks.source_id
			JOIN tape_to_tracks ON tape_to_tracks.track_id = tracks.id
			JOIN tapes ON tapes.id = tape_to_tracks.tape_id
			LEFT JOIN track_listens ON track_listens.track_id = tracks.id
			WHERE tapes.id = '%s' AND tapes.type = '%s'
			ORDER BY album_track_index ASC
		`,
		albumId.String(),
		TAPE_TYPE_ALBUM,
	)

	result := []SubsonicTrackItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *TrackStorage) GetSubsonicTracksByPlaylist(playlistId uuid.UUID) ([]SubsonicTrackItem, error) {
	query := fmt.Sprintf(
		`
			WITH filtered_tracks AS (
				SELECT
					tracks.id AS id,
					sources.thumbnail_id AS thumbnail_id,
					tape_to_tracks.list_index AS playlist_track_index,
					tracks.artist AS artist,
					tracks.title AS title,
					(tracks.end_offset_ms - tracks.start_offset_ms) / 1000 AS duration_sec,
					track_listens.listen_count AS play_count
				FROM tracks
				JOIN sources ON sources.id = tracks.source_id
				JOIN tape_to_tracks ON tape_to_tracks.track_id = tracks.id
				JOIN tapes playlists ON playlists.id = tape_to_tracks.tape_id AND playlists.type = '%s'
				LEFT JOIN track_listens ON track_listens.track_id = tracks.id
				WHERE playlists.id = '%s'
			)
			SELECT
				enriched_tracks.*
			FROM (
				SELECT
					filtered_tracks.*,
					albums.id AS album_id,
					albums.thumbnail_id AS album_thumbnail_id,
					albums.name AS album,
					tape_to_tracks.list_index AS album_track_index,
					row_number() OVER (PARTITION BY filtered_tracks.id, filtered_tracks.playlist_track_index ORDER BY albums.created_at ASC NULLS LAST) AS rank
				FROM filtered_tracks
				LEFT JOIN tape_to_tracks ON tape_to_tracks.track_id = filtered_tracks.id
				LEFT JOIN tapes albums ON albums.id = tape_to_tracks.tape_id AND albums.type = '%s'
			) enriched_tracks
			WHERE enriched_tracks.rank = 1
			ORDER BY enriched_tracks.playlist_track_index ASC
		`,
		TAPE_TYPE_PLAYLIST,
		playlistId.String(),
		TAPE_TYPE_ALBUM,
	)

	result := []SubsonicTrackItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *TrackStorage) getSubsonicTracks(count int, offset int, filter string, order string) ([]SubsonicTrackItem, error) {
	if filter == "" {
		filter = "1 = 1"
	}

	query := fmt.Sprintf(
		`
			WITH filtered_tracks AS (
				SELECT
					tracks.id AS id,
					sources.thumbnail_id AS thumbnail_id,
					tracks.artist AS artist,
					tracks.title AS title,
					(tracks.end_offset_ms - tracks.start_offset_ms) / 1000 AS duration_sec,
					track_listens.listen_count AS play_count
				FROM tracks
				JOIN sources ON sources.id = tracks.source_id
				LEFT JOIN track_listens ON track_listens.track_id = tracks.id
			)
			SELECT
				enriched_tracks.*
			FROM (
				SELECT
					filtered_tracks.*,
					albums.id AS album_id,
					albums.thumbnail_id AS album_thumbnail_id,
					albums.name AS album,
					tape_to_tracks.list_index AS album_track_index,
					albums.released_at AS album_release_date,
					row_number() OVER (PARTITION BY filtered_tracks.id ORDER BY albums.created_at ASC NULLS LAST) AS rank
				FROM filtered_tracks
				LEFT JOIN tape_to_tracks ON tape_to_tracks.track_id = filtered_tracks.id
				LEFT JOIN tapes albums ON albums.id = tape_to_tracks.tape_id AND albums.type = '%s'
			) enriched_tracks
			WHERE enriched_tracks.rank = 1 AND %s
			ORDER BY %s
			LIMIT %d
			OFFSET %d
		`,
		TAPE_TYPE_ALBUM,
		filter,
		order,
		count,
		offset,
	)

	result := []SubsonicTrackItem{}
	return result, storage.db.Raw(query).Find(&result).Error
}
