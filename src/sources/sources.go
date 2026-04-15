package sources

import (
	"errors"
	"fmt"
	"strings"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SourceStorage struct {
	db *gorm.DB
}

func newSourceStorage(db *gorm.DB) *SourceStorage {
	return &SourceStorage{db: db}
}

func (store *SourceStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&Source{}, &SourceHierarchy{})
}

func (store *SourceStorage) Upsert(source Source) (Source, error) {
	query := `
		INSERT INTO sources (id, extractor_key, extracted_id, url, title, uploader, uploader_id, album_artist, album_title, album_index, track_artist, track_title, duration_ms, uploaded_at, release_date, artwork_id, management_policy, created_at, updated_at)
		VALUES (@id, @extractorKey, @extractedId, @url, @title, @uploader, @uploaderId, @albumArtist, @albumTitle, @albumIndex, @trackArtist, @trackTitle, @durationMs, @uploadedAt, @releaseDate, @artworkId, @managementPolicy, @createdAt, @updatedAt)
		ON CONFLICT (url) DO UPDATE
		SET
			extractor_key = excluded.extractor_key,
			extracted_id = excluded.extracted_id,
			url = excluded.url,
			title = excluded.title,
			uploader = excluded.uploader,
			uploader_id = excluded.uploader_id,
			album_artist = excluded.album_artist,
			album_title = excluded.album_title,
			album_index = excluded.album_index,
			track_artist = excluded.track_artist,
			track_title = excluded.track_title,
			duration_ms = excluded.duration_ms,
			uploaded_at = excluded.uploaded_at,
			release_date = excluded.release_date,
			artwork_id = excluded.artwork_id,
			management_policy = excluded.management_policy,
			updated_at = excluded.updated_at
		RETURNING *
	`
	params := map[string]any{
		"id":               source.Id,
		"extractorKey":     source.ExtractorKey,
		"extractedId":      source.ExtractedId,
		"url":              source.Url,
		"title":            source.Title,
		"uploader":         source.Uploader,
		"uploaderId":       source.UploaderId,
		"albumArtist":      source.AlbumArtist,
		"albumTitle":       source.AlbumTitle,
		"albumIndex":       source.AlbumIndex,
		"trackArtist":      source.TrackArtist,
		"trackTitle":       source.TrackTitle,
		"durationMs":       source.DurationMs,
		"uploadedAt":       source.UploadedAt,
		"releaseDate":      source.ReleaseDate,
		"artworkId":        source.ArtworkId,
		"managementPolicy": source.ManagementPolicy,
		"createdAt":        source.CreatedAt,
		"updatedAt":        source.UpdatedAt,
	}

	return source, store.db.Raw(query, params).First(&source).Error
}

func (store *SourceStorage) UpdateHierarchy(parentId uuid.UUID, childIds []uuid.UUID) error {
	sql := `
		INSERT INTO source_hierarchies (parent_id, child_id, list_index)
		VALUES (@parentId, @childId, @listIndex)
		ON CONFLICT DO UPDATE
		SET list_index = excluded.list_index
	`

	return store.db.Transaction(func(tx *gorm.DB) error {
		for i, childId := range childIds {
			params := map[string]any{
				"parentId":  parentId,
				"childId":   childId,
				"listIndex": i,
			}

			err := store.db.Exec(sql, params).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (store *SourceStorage) GetHierarchy(id uuid.UUID) ([]SourceForHierarchy, error) {
	query := `
		WITH RECURSIVE all_sources (id) AS (
			VALUES (@id)
			UNION
			SELECT ids.value
			FROM source_hierarchies, json_each(json_array(source_hierarchies.parent_id, source_hierarchies.child_id)) ids
			JOIN all_sources ON all_sources.id IN (source_hierarchies.parent_id, source_hierarchies.child_id)
		)
		SELECT
			sources.id AS id,
			source_hierarchies.parent_id AS parent_id,
			coalesce(source_hierarchies.list_index, -1) AS list_index,
			sources.url AS url,
			sources.title AS title,
			sources.uploader AS uploader,
			sources.artwork_id AS artwork_id
		FROM sources
		JOIN all_sources ON sources.id = all_sources.id
		LEFT JOIN source_hierarchies ON sources.id = source_hierarchies.child_id
	`
	params := map[string]any{
		"id": id,
	}

	result := []SourceForHierarchy{}
	return result, store.db.Raw(query, params).Find(&result).Error
}

func (store *SourceStorage) GetListForApi(managementPolicies []SourceManagementPolicy) ([]Source, error) {
	conditions := []string{"1 = 1"}
	params := map[string]any{}

	if len(managementPolicies) > 0 {
		conditions = append(conditions, "management_policy IN @managementPolicies")
		params["managementPolicies"] = managementPolicies
	}

	sql := fmt.Sprintf(
		`
			SELECT *
			FROM sources
			WHERE %s
			ORDER BY created_at DESC, uploaded_at DESC, album_index DESC, id DESC
		`,
		strings.Join(conditions, " AND "),
	)

	result := []Source{}
	if len(params) > 0 {
		return result, store.db.Raw(sql, params).Find(&result).Error
	} else {
		return result, store.db.Raw(sql).Find(&result).Error
	}
}

func (store *SourceStorage) GetById(id uuid.UUID) (Source, error) {
	sql := `
		SELECT *
		FROM sources
		WHERE id = @id
		LIMIT 1
	`
	params := map[string]any{
		"id": id,
	}

	result := Source{}
	err := store.db.Raw(sql, params).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, model.ErrNotFound
	} else {
		return result, err
	}
}

func (store *SourceStorage) FindByUrl(url string) (*Source, error) {
	result := Source{}
	if err := store.db.Where("url = ?", url).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &result, nil
}

func (store *SourceStorage) GetManagementPolicyById(id uuid.UUID) (SourceManagementPolicy, error) {
	result := SOURCE_MANAGEMENT_POLICY_MANUAL
	return result, store.db.Raw("SELECT management_policy FROM sources WHERE id = ?", id).Take(&result).Error
}

func (store *SourceStorage) SetManagementPolicyById(id uuid.UUID, managementPolicy SourceManagementPolicy) error {
	return store.db.Exec("UPDATE sources SET management_policy = ? WHERE id = ?", managementPolicy, id).Error
}

func (store *SourceStorage) FindNextForDownload() (*Source, error) {
	sql := `
		SELECT sources.*
		FROM sources
		LEFT JOIN source_files ON source_files.source_id = sources.id
		WHERE
			sources.duration_ms > 0
			AND source_files.id IS NULL
			AND EXISTS (
				SELECT 1
				FROM source_tracks
				JOIN tape_to_tracks ON tape_to_tracks.track_id = source_tracks.id
				WHERE source_tracks.source_id = sources.id
				LIMIT 1
			)
		ORDER BY random()
		LIMIT 1
	`

	result := Source{}
	if err := store.db.Raw(sql).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &result, nil
}
