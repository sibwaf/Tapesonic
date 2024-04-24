package tasks

import (
	"fmt"
	"log/slog"
	"tapesonic/config"
	"tapesonic/storage"

	"github.com/robfig/cron/v3"
)

type ImportQueueTaskHandler struct {
	importQueueStorage *storage.ImportQueueStorage
	importer           *storage.Importer

	importConfig config.BackgroundTaskConfig
}

func NewImportQueueTaskHandler(
	importQueueStorage *storage.ImportQueueStorage,
	importer *storage.Importer,

	importConfig config.BackgroundTaskConfig,
) *ImportQueueTaskHandler {
	return &ImportQueueTaskHandler{
		importQueueStorage: importQueueStorage,
		importer:           importer,

		importConfig: importConfig,
	}
}

func (h *ImportQueueTaskHandler) RegisterSchedules(cron *cron.Cron) error {
	_, err := cron.AddFunc(h.importConfig.Cron, h.onImportCron)
	return err
}

func (h *ImportQueueTaskHandler) onImportCron() {
	queueItem, err := h.importQueueStorage.FetchNext(h.importConfig.Cooldown)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to fetch next import queue item: %s", err.Error()))
		return
	}

	if queueItem == nil {
		slog.Debug("No pending items in import queue")
		return
	}

	slog.Info(fmt.Sprintf("Importing import queue item %s (%s)", queueItem.Id, queueItem.Url))

	tape, err := h.importer.ImportTape(queueItem.Url, "ba")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to import queue item %s (%s): %s", queueItem.Id, queueItem.Url, err.Error()))
		if err = h.importQueueStorage.Fail(queueItem.Id); err != nil {
			slog.Error(fmt.Sprintf("Failed to mark failure for import queue item %s (%s): %s", queueItem.Id, queueItem.Url, err.Error()))
		}
		return
	}

	err = h.importQueueStorage.Complete(queueItem.Id, tape.Id)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to bind complete import queue item %s (%s): %s", queueItem.Id, queueItem.Url, err.Error()))
		return
	}

	slog.Info(fmt.Sprintf("Successfully imported import queue item %s (%s)", queueItem.Id, queueItem.Url))
}
