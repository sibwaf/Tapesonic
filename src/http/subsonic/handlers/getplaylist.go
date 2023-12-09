package handlers

import (
	"fmt"
	"net/http"

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

	playlist, err := h.dataStorage.GetPlaylistWithTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for _, track := range playlist.Tracks {
		lengthMs := track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs

		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.TapeTrack.Id),
			false,
			track.TapeTrack.Artist,
			track.TapeTrack.Title,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
	}

	response := responses.NewOkResponse()
	response.Playlist = responses.NewSubsonicPlaylist(
		fmt.Sprint(playlist.Id),
		playlist.Name,
		len(playlist.Tracks),
		totalLengthMs/1000,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	)
	response.Playlist.CoverArt = fmt.Sprint(playlist.Id)
	response.Playlist.Entry = tracks

	return response, nil
}
