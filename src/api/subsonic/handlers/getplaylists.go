package handlers

import (
	"net/http"
	"time"

	"tapesonic/api/subsonic/responses"
	"tapesonic/storage"
)

type getPlaylistsHandler struct {
	storage *storage.Storage
}

func NewGetPlaylistsHandler(
	storage *storage.Storage,
) *getPlaylistsHandler {
	return &getPlaylistsHandler{
		storage: storage,
	}
}

func (h *getPlaylistsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	tapes, err := h.storage.GetTapes()
	if err != nil {
		return nil, err
	}

	playlists := []responses.SubsonicPlaylist{}
	for _, tape := range tapes {
		totalLengthMs := 0
		for _, track := range tape.Tracks {
			totalLengthMs += track.LengthMs
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
