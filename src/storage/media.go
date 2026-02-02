package storage

import (
	"fmt"
	"path"
	"tapesonic/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaStorage struct {
	db *DbHelper

	dir string
}

type CoverDescriptor struct {
	Path   string
	Format string
}

func NewMediaStorage(
	db *gorm.DB,
	dir string,
) *MediaStorage {
	return &MediaStorage{
		db: NewDbHelper(db),

		dir: dir,
	}
}

func (ms *MediaStorage) GetTrackSources(trackId uuid.UUID) (model.TrackSourceDescriptor, error) {
	query := fmt.Sprintf(
		`
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
			WHERE source_tracks.id = '%s'
		`,
		trackId.String(),
	)

	sourceDescriptor := model.TrackSourceDescriptor{}
	if err := ms.db.Raw(query).Find(&sourceDescriptor).Error; err != nil {
		return sourceDescriptor, err
	}

	if sourceDescriptor.LocalPath != "" {
		sourceDescriptor.LocalPath = path.Join(ms.dir, sourceDescriptor.LocalPath)
	}

	return sourceDescriptor, nil
}
