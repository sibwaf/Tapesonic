package handlers

import (
	"context"
	"net/http"

	"tapesonic/http/admin/responses"
	"tapesonic/logic"
	"tapesonic/model"
)

type GetListSourceRs struct {
	Source responses.ListSourceRs
	File   *responses.SourceFileRs
}

type sourcesHandler struct {
	service *logic.SourceService
}

func NewSourcesHandler(
	service *logic.SourceService,
) *sourcesHandler {
	return &sourcesHandler{
		service: service,
	}
}

func (h *sourcesHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *sourcesHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		managementPolicies := []string{}

		for key, values := range r.URL.Query() {
			if key == "managementPolicy" {
				managementPolicies = append(managementPolicies, values...)
			}
		}

		sources, err := h.service.GetListForApi(managementPolicies)
		if err != nil {
			return nil, err
		}

		response := []GetListSourceRs{}
		for _, item := range sources {
			itemRs := GetListSourceRs{
				Source: responses.SourceToListSourceRs(item.Source),
			}

			if item.File != nil {
				file := responses.SourceFileToSourceFileRs(*item.File)
				itemRs.File = &file
			}

			response = append(response, itemRs)
		}

		return response, nil
	case http.MethodPost:
		url := r.URL.Query().Get("url")
		if url == "" {
			resp := responses.NewResponse("`url` query parameter missing") // todo
			return &resp, nil
		}

		source, err := h.service.AddSource(context.Background(), url, model.SOURCE_MANAGEMENT_POLICY_MANUAL)
		if err != nil {
			return nil, err
		}

		return responses.SourceToFullSourceRs(source), nil
	default:
		return nil, http.ErrNotSupported
	}
}
