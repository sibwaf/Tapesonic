package sources

import (
	"fmt"
	"log/slog"
	"tapesonic/logic"
	"tapesonic/scheduling"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type SourcesModule struct {
	SourceService *SourceService
	ImportService *ImportService

	sources *SourceStorage
	tracks  *TrackStorage
	files   *FileStorage

	downloads *downloadService

	sourceDownloadCron cron.Schedule
}

func NewSourcesModule(
	db *gorm.DB,
	ytdlp *logic.YtdlpService,
	thumbnails *logic.ThumbnailService,
	sourceDownloadCron cron.Schedule,
	mediaDir string,
) *SourcesModule {
	sources := newSourceStorage(db)
	tracks := newTrackStorage(db)
	files := newFileStorage(db)

	normalizer := NewTrackNormalizer()

	downloads := newDownloadService(sources, files, ytdlp, mediaDir)

	return &SourcesModule{
		SourceService: newSourceService(sources, tracks, files, mediaDir),
		ImportService: newImportService(sources, tracks, normalizer, ytdlp, thumbnails),

		sources: sources,
		tracks:  tracks,
		files:   files,

		downloads: downloads,

		sourceDownloadCron: sourceDownloadCron,
	}
}

func (module *SourcesModule) PrepareDatabase() error {
	if err := module.sources.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.tracks.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.files.PrepareDatabase(); err != nil {
		return err
	}

	return nil
}

func (module *SourcesModule) RegisterSchedules(scheduler *scheduling.TaskScheduler) {
	if module.sourceDownloadCron != nil {
		if _, err := scheduler.RegisterTask(TASK_SOURCES_FIND_SOURCE_FOR_DOWNLOAD, nil, time.Now()); err != nil {
			slog.Error(fmt.Sprintf("Failed to register source download task: %s", err.Error()))
		}

		retryDelay := 1 * time.Minute // todo: config

		scheduling.SubscribeTaskRunner(
			scheduler,
			TASK_SOURCES_FIND_SOURCE_FOR_DOWNLOAD,
			retryDelay,
			func(scheduler *scheduling.TaskScheduler, task scheduling.ScheduledTask, parameters any) error {
				if err := module.downloads.DownloadNextPending(); err != nil {
					return err
				}

				return scheduler.RescheduleTask(task.Id, module.sourceDownloadCron.Next(time.Now()))
			},
		)
	}
}
