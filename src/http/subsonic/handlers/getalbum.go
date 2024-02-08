package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type getAlbumHandler struct {
	albumStorage *storage.AlbumStorage
}

func NewGetAlbumHandler(
	albumStorage *storage.AlbumStorage,
) *getAlbumHandler {
	return &getAlbumHandler{
		albumStorage: albumStorage,
	}
}

func (h *getAlbumHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	rawId := r.URL.Query().Get("id")
	if rawId == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	album, err := h.albumStorage.GetAlbumWithTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for index, track := range album.Tracks {
		lengthMs := track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs

		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.TapeTrack.Id),
			false,
			track.TapeTrack.Artist,
			track.TapeTrack.Title,
			index+1,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
	}

	response := responses.NewOkResponse()
	response.Album = responses.NewAlbumId3(
		fmt.Sprint(album.Id),
		album.Name,
		album.Artist,
		"album/"+fmt.Sprint(album.Id),
		len(album.Tracks),
		totalLengthMs/1000,
		album.CreatedAt,
	)
	response.Album.Song = tracks

	return response, nil
}
