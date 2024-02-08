package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
)

type getPlaylistsHandler struct {
	playlistStorage *storage.PlaylistStorage
}

func NewGetPlaylistsHandler(
	playlistStorage *storage.PlaylistStorage,
) *getPlaylistsHandler {
	return &getPlaylistsHandler{
		playlistStorage: playlistStorage,
	}
}

func (h *getPlaylistsHandler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	playlists, err := h.playlistStorage.GetAllPlaylists()
	if err != nil {
		return nil, err
	}

	responsePlaylists := []responses.SubsonicPlaylist{}
	for _, playlist := range playlists {
		totalLengthMs := 0
		for _, track := range playlist.Tracks {
			totalLengthMs += track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs
		}

		responsePlaylist := responses.NewSubsonicPlaylist(
			fmt.Sprint(playlist.Id),
			playlist.Name,
			len(playlist.Tracks),
			totalLengthMs/1000,
			playlist.CreatedAt,
			playlist.UpdatedAt,
		)
		responsePlaylist.CoverArt = "playlist/" + fmt.Sprint(responsePlaylist.Id)

		responsePlaylists = append(responsePlaylists, *responsePlaylist)
	}

	response := responses.NewOkResponse()
	response.Playlists = responses.NewSubsonicPlaylists(responsePlaylists)

	return response, nil
}
