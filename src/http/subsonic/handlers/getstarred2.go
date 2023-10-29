package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
)

type getStarred2Handler struct {
}

func NewGetStarred2Handler() *getStarred2Handler {
	return &getStarred2Handler{}
}

func (h *getStarred2Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	// todo
	response := responses.NewOkResponse()
	response.Starred2 = responses.NewStarred2()
	return response, nil
}
