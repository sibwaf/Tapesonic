package handlers

import (
	"errors"
	"net/http"

	"tapesonic/storage"
)

type importQueueHandler struct {
	queue *storage.ImportQueueStorage
}

func NewImportQueueHandler(
	queue *storage.ImportQueueStorage,
) *importQueueHandler {
	return &importQueueHandler{
		queue: queue,
	}
}

func (h *importQueueHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *importQueueHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		return h.queue.GetAllEnqueued()
	case http.MethodPost:
		url := r.URL.Query().Get("url")
		if url == "" {
			return nil, errors.New("no url provided")
		}
		return h.queue.Enqueue(url)
	default:
		return nil, http.ErrNotSupported
	}
}
