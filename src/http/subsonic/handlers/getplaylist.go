package handlers

import (
	"fmt"
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type getPlaylistHandler struct {
	dataStorage *storage.DataStorage
}

func NewGetPlaylistHandler(
	dataStorage *storage.DataStorage,
) *getPlaylistHandler {
	return &getPlaylistHandler{
		dataStorage: dataStorage,
	}
}

func (h *getPlaylistHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	rawId := r.URL.Query().Get("id")
	if rawId == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	tape, err := h.dataStorage.GetTapeWithTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for _, track := range tape.Tracks {
		lengthMs := track.EndOffsetMs - track.StartOffsetMs

		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.Id),
			false,
			track.Artist,
			track.Title,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
	}

	response := responses.NewOkResponse()
	response.Playlist = responses.NewSubsonicPlaylist(
		fmt.Sprint(tape.Id),
		tape.Name,
		len(tape.Tracks),
		totalLengthMs/1000,
		time.Now(),
		time.Now(),
	)
	response.Playlist.CoverArt = fmt.Sprint(tape.Id)
	response.Playlist.Owner = tape.AuthorName
	response.Playlist.Entry = tracks

	return response, nil
}
