package sources

import (
	"fmt"
	"log/slog"
	"tapesonic/logic"
	"tapesonic/scheduling"
	"tapesonic/storage"
	"time"

	"github.com/robfig/cron/v3"
)

// todo: move source-related things here

type SourcesModule struct {
	files              *logic.SourceFileService
	sources            *storage.SourceStorage
	sourceDownloadCron cron.Schedule
}

func NewSourcesModule(
	files *logic.SourceFileService,
	sources *storage.SourceStorage,
	sourceDownloadCron cron.Schedule,
) *SourcesModule {
	return &SourcesModule{
		files:              files,
		sources:            sources,
		sourceDownloadCron: sourceDownloadCron,
	}
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
				// todo: rework into task per source?

				source, err := module.sources.FindNextForDownload()
				if err != nil {
					return err
				}

				if source != nil {
					_, err := module.files.DownloadIfMissingFor(source.Id)
					if err != nil {
						return err
					}
				}

				return scheduler.RescheduleTask(task.Id, module.sourceDownloadCron.Next(time.Now()))
			},
		)
	}
}
