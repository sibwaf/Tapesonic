package handlers

import (
	"fmt"
	"net/http"
	"time"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
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
	id := r.URL.Query().Get("id")
	if id == "" {
		return responses.NewFailedResponse(responses.ERROR_CODE_PARAMETER_MISSING, "no playlist id"), nil
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
			fmt.Sprintf("%s/%d", tape.Id, track.TapeTrackIndex),
			false,
			track.Title,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
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
	response.Playlist.Owner = tape.AuthorName
	response.Playlist.Entry = tracks

	return response, nil
}
