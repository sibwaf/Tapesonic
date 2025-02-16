package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/logic"

	"github.com/google/uuid"
)

type ListThumbnailRs struct {
	Id uuid.UUID
}

type thumbnailsHandler struct {
	thumbnails *logic.ThumbnailService
}

func NewThumbnailsHandler(
	thumbnails *logic.ThumbnailService,
) *thumbnailsHandler {
	return &thumbnailsHandler{
		thumbnails: thumbnails,
	}
}

func (h *thumbnailsHandler) Methods() []string {
	return []string{http.MethodGet}
}

func (h *thumbnailsHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		sourceIds := []uuid.UUID{}
		for _, sourceId := range r.URL.Query()["sourceId"] {
			parsed, err := uuid.Parse(sourceId)
			if err != nil {
				return nil, fmt.Errorf("invalid sourceId: %s", sourceId)
			}

			sourceIds = append(sourceIds, parsed)
		}

		thumbnails, err := h.thumbnails.GetListForApi(sourceIds)
		if err != nil {
			return nil, err
		}

		response := []ListThumbnailRs{}
		for _, item := range thumbnails {
			itemRs := ListThumbnailRs{
				Id: item.Id,
			}

			response = append(response, itemRs)
		}

		return response, nil
	default:
		return nil, http.ErrNotSupported
	}
}
