package handlers

import (
	"fmt"
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
)

type getPlaylistHandler struct {
	storage *storage.Storage
}

func NewGetPlaylistHandler(
	storage *storage.Storage,
) *getPlaylistHandler {
	return &getPlaylistHandler{
		storage: storage,
	}
}

func (h *getPlaylistHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewFailedResponse(responses.ERROR_CODE_PARAMETER_MISSING, "no playlist id"), nil
	}

	tape, err := h.storage.GetTape(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for _, track := range tape.Tracks {
		trackResponse := responses.NewSubsonicChild(
			fmt.Sprintf("%s/%d", tape.Id, track.Index),
			false,
			track.Name,
			track.LengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += track.LengthMs
	}

	response := responses.NewOkResponse()
	response.Playlist = responses.NewSubsonicPlaylist(
		tape.Id,
		tape.Name,
		len(tape.Tracks),
		totalLengthMs/1000,
		time.Now(),
		time.Now(),
	)
	response.Playlist.CoverArt = tape.Id
	response.Playlist.Owner = tape.Author
	response.Playlist.Entry = tracks

	return response, nil
}
