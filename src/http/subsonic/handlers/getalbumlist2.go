package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getAlbumList2Handler struct {
}

func NewGetAlbumList2Handler() *getAlbumList2Handler {
	return &getAlbumList2Handler{}
}

func (h *getAlbumList2Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.AlbumList2 = responses.NewAlbumList2([]responses.AlbumId3{})
	return response, nil
}
