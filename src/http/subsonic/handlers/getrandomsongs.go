package handlers

import (
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/logic"
	"tapesonic/util"
)

type getRandomSongsHandler struct {
	subsonic logic.SubsonicService
}

func NewGetRandomSongsHandler(subsonic logic.SubsonicService) *getRandomSongsHandler {
	return &getRandomSongsHandler{
		subsonic: subsonic,
	}
}

func (h *getRandomSongsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	size := util.StringToIntOrDefault(r.URL.Query().Get("size"), 10)
	genre := r.URL.Query().Get("genre")
	fromYear := util.StringToIntOrNull(r.URL.Query().Get("fromYear"))
	toYear := util.StringToIntOrNull(r.URL.Query().Get("toYear"))

	songs, err := h.subsonic.GetRandomSongs(size, genre, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	response := responses.NewOkResponse()
	response.RandomSongs = songs
	return response, nil
}
