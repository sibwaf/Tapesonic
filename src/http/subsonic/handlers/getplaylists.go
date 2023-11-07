package handlers

import (
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
)

type getPlaylistsHandler struct {
	dataStorage *storage.DataStorage
}

func NewGetPlaylistsHandler(
	dataStorage *storage.DataStorage,
) *getPlaylistsHandler {
	return &getPlaylistsHandler{
		dataStorage: dataStorage,
	}
}

func (h *getPlaylistsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	tapes, err := h.dataStorage.GetAllTapes()
	if err != nil {
		return nil, err
	}

	playlists := []responses.SubsonicPlaylist{}
	for _, tape := range tapes {
		totalLengthMs := 0
		for _, track := range tape.Tracks {
			totalLengthMs += track.EndOffsetMs - track.StartOffsetMs
		}

		playlist := responses.NewSubsonicPlaylist(
			tape.Id,
			tape.Name,
			len(tape.Tracks),
			totalLengthMs/1000,
			time.Now(),
			time.Now(),
		)
		playlist.CoverArt = tape.Id

		playlists = append(playlists, *playlist)
	}

	response := responses.NewOkResponse()
	response.Playlists = responses.NewSubsonicPlaylists(playlists)

	return response, nil
}
