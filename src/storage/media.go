package storage

import (
	"fmt"
	"path"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaStorage struct {
	db *DbHelper

	dir string

	tapeStorage     *TapeStorage
	playlistStorage *PlaylistStorage
	albumStorage    *AlbumStorage
}

type TrackSourceDescriptor struct {
	LocalPath   string
	LocalFormat string
	LocalCodec  string

	RemoteUrl        string
	SourceDurationMs int64

	StartOffsetMs int64
	EndOffsetMs   int64
}

type CoverDescriptor struct {
	Path   string
	Format string
}

func NewMediaStorage(
	db *gorm.DB,
	dir string,
	tapeStorage *TapeStorage,
	playlistStorage *PlaylistStorage,
	albumStorage *AlbumStorage,
) *MediaStorage {
	return &MediaStorage{
		db: NewDbHelper(db),

		dir: dir,

		tapeStorage:     tapeStorage,
		playlistStorage: playlistStorage,
		albumStorage:    albumStorage,
	}
}

func (ms *MediaStorage) GetTrackSources(trackId uuid.UUID) (TrackSourceDescriptor, error) {
	query := fmt.Sprintf(
		`
			SELECT
				source_files.media_path AS local_path,
				source_files.format AS local_format,
				source_files.codec AS local_codec,
				sources.url AS remote_url,
				sources.duration_ms AS source_duration_ms,
				tracks.start_offset_ms AS start_offset_ms,
				tracks.end_offset_ms AS end_offset_ms
			FROM tracks
			JOIN sources ON sources.id = tracks.source_id
			LEFT JOIN source_files ON source_files.source_id = sources.id
			WHERE tracks.id = '%s'
		`,
		trackId.String(),
	)

	sourceDescriptor := TrackSourceDescriptor{}
	if err := ms.db.Raw(query).Find(&sourceDescriptor).Error; err != nil {
		return sourceDescriptor, err
	}

	if sourceDescriptor.LocalPath != "" {
		sourceDescriptor.LocalPath = path.Join(ms.dir, sourceDescriptor.LocalPath)
	}

	return sourceDescriptor, nil
}
