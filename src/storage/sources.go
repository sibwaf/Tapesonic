package storage

import (
	"errors"
	"fmt"
	"strings"
	"tapesonic/model"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Source struct {
	Id uuid.UUID

	ExtractorKey string
	ExtractedId  string
	Url          string `gorm:"uniqueIndex"`

	Title      string
	Uploader   string
	UploaderId string

	AlbumArtist string
	AlbumTitle  string
	AlbumIndex  int
	TrackArtist string
	TrackTitle  string
	DurationMs  int64

	UploadedAt  util.TimestampWrapper
	ReleaseDate *util.TimestampWrapper

	ThumbnailId *uuid.UUID
	Thumbnail   *Thumbnail

	ManagementPolicy model.SourceManagementPolicy

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type SourceHierarchy struct {
	ParentId uuid.UUID `gorm:"primaryKey"`
	Parent   Source

	ChildId uuid.UUID `gorm:"primaryKey"`
	Child   Source

	ListIndex int
}

type SourceStorage struct {
	db *DbHelper
}

func NewSourceStorage(db *gorm.DB) (*SourceStorage, error) {
	if err := db.AutoMigrate(&Source{}, &SourceHierarchy{}); err != nil {
		return nil, err
	}

	return &SourceStorage{db: NewDbHelper(db)}, nil
}

func (storage *SourceStorage) Upsert(source Source) (Source, error) {
	query := `
		INSERT INTO sources (id, extractor_key, extracted_id, url, title, uploader, uploader_id, album_artist, album_title, album_index, track_artist, track_title, duration_ms, uploaded_at, release_date, thumbnail_id, management_policy, created_at, updated_at)
		VALUES (@id, @extractorKey, @extractedId, @url, @title, @uploader, @uploaderId, @albumArtist, @albumTitle, @albumIndex, @trackArtist, @trackTitle, @durationMs, @uploadedAt, @releaseDate, @thumbnailId, @managementPolicy, @createdAt, @updatedAt)
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
			thumbnail_id = excluded.thumbnail_id,
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
		"thumbnailId":      source.ThumbnailId,
		"managementPolicy": source.ManagementPolicy,
		"createdAt":        source.CreatedAt,
		"updatedAt":        source.UpdatedAt,
	}

	return source, storage.db.Raw(query, params).First(&source).Error
}

func (storage *SourceStorage) UpdateHierarchy(parentId uuid.UUID, childIds []uuid.UUID) error {
	return storage.db.Transaction(func(tx *gorm.DB) error {
		items := []SourceHierarchy{}
		for i, childId := range childIds {
			items = append(items, SourceHierarchy{
				ParentId:  parentId,
				ChildId:   childId,
				ListIndex: i,
			})
		}

		if err := tx.Where("parent_id = ?", parentId.String()).Delete(&SourceHierarchy{}).Error; err != nil {
			return err
		}
		if err := tx.Create(items).Error; err != nil {
			return err
		}

		return nil
	})
}

func (storage *SourceStorage) GetHierarchy(id uuid.UUID) ([]SourceForHierarchy, error) {
	query := fmt.Sprintf(
		`
			WITH RECURSIVE all_sources (id) AS (
				VALUES ('%s')
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
				sources.thumbnail_id AS thumbnail_id
			FROM sources
			JOIN all_sources ON sources.id = all_sources.id
			LEFT JOIN source_hierarchies ON sources.id = source_hierarchies.child_id
		`,
		id,
	)

	result := []SourceForHierarchy{}
	return result, storage.db.Raw(query).Find(&result).Error
}

func (storage *SourceStorage) GetListForApi(managementPolicies []model.SourceManagementPolicy) ([]Source, error) {
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
		return result, storage.db.Raw(sql, params).Find(&result).Error
	} else {
		return result, storage.db.Raw(sql).Find(&result).Error
	}
}

func (storage *SourceStorage) GetById(id uuid.UUID) (Source, error) {
	result := Source{Id: id}
	return result, storage.db.Take(&result).Error
}

func (storage *SourceStorage) FindByUrl(url string) (*Source, error) {
	result := Source{}
	if err := storage.db.Where("url = ?", url).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &result, nil
}

func (storage *SourceStorage) GetManagementPolicyById(id uuid.UUID) (model.SourceManagementPolicy, error) {
	result := model.SOURCE_MANAGEMENT_POLICY_MANUAL
	return result, storage.db.Raw("SELECT management_policy FROM sources WHERE id = ?", id).Take(&result).Error
}

func (storage *SourceStorage) SetManagementPolicyById(id uuid.UUID, managementPolicy model.SourceManagementPolicy) error {
	return storage.db.Exec("UPDATE sources SET management_policy = ? WHERE id = ?", managementPolicy, id).Error
}

func (storage *SourceStorage) FindNextForDownload() (*Source, error) {
	sql := `
		SELECT sources.*
		FROM sources
		LEFT JOIN source_files ON source_files.source_id = sources.id
		WHERE
			sources.duration_ms > 0
			AND source_files.id IS NULL
			AND EXISTS (
				SELECT 1
				FROM tracks
				JOIN tape_to_tracks ON tape_to_tracks.track_id = tracks.id
				WHERE tracks.source_id = sources.id
				LIMIT 1
			)
		ORDER BY random()
		LIMIT 1
	`

	result := Source{}
	if err := storage.db.Raw(sql).Take(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &result, nil
}
