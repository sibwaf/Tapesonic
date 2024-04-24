package handlers

import (
	"net/http"

	"tapesonic/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type importQueueItemHandler struct {
	queue *storage.ImportQueueStorage
}

func NewImportQueueItemHandler(
	queue *storage.ImportQueueStorage,
) *importQueueItemHandler {
	return &importQueueItemHandler{
		queue: queue,
	}
}

func (h *importQueueItemHandler) Methods() []string {
	return []string{http.MethodDelete}
}

func (h *importQueueItemHandler) Handle(r *http.Request) (any, error) {
	rawId := mux.Vars(r)["itemId"]
	id, idErr := uuid.Parse(rawId)
	if idErr != nil {
		return nil, idErr
	}

	switch r.Method {
	case http.MethodDelete:
		return nil, h.queue.Delete(id)
	default:
		return nil, http.ErrNotSupported
	}
}
