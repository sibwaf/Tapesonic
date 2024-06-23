package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
	"tapesonic/util"
)

type getAlbumList2Handler struct {
	subsonic logic.SubsonicService
}

func NewGetAlbumList2Handler(subsonic logic.SubsonicService) *getAlbumList2Handler {
	return &getAlbumList2Handler{
		subsonic: subsonic,
	}
}

func (h *getAlbumList2Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	listType := r.URL.Query().Get("type")
	if listType == "" {
		return responses.NewParameterMissingResponse("type"), nil
	}

	size := util.StringToIntOrDefault(r.URL.Query().Get("size"), 10)
	offset := util.StringToIntOrDefault(r.URL.Query().Get("offset"), 0)
	fromYear := util.StringToIntOrNull(r.URL.Query().Get("fromYear"))
	toYear := util.StringToIntOrNull(r.URL.Query().Get("toYear"))

	albums, err := h.subsonic.GetAlbumList2(listType, size, offset, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.AlbumList2 = albums
	return response, nil
}
