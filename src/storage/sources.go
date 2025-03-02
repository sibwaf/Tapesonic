package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	UploadedAt  time.Time
	ReleaseDate *time.Time

	ThumbnailId *uuid.UUID
	Thumbnail   *Thumbnail

	CreatedAt time.Time
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
	if source.Id == uuid.Nil {
		source.Id = uuid.New()
	}

	return source, storage.db.Clauses(
		clause.OnConflict{Columns: []clause.Column{{Name: "url"}}, UpdateAll: true},
		clause.Returning{},
	).Create(&source).Error
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

func (storage *SourceStorage) GetAll() ([]Source, error) {
	result := []Source{}
	return result, storage.db.Order("created_at DESC, uploaded_at DESC, album_index DESC, id DESC").Find(&result).Error
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
