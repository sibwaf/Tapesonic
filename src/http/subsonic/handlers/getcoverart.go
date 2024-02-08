package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"tapesonic/http/subsonic/responses"
	"tapesonic/http/util"
	"tapesonic/storage"

	"github.com/google/uuid"
)

type getCoverArtHandler struct {
	mediaStorage *storage.MediaStorage
}

func NewGetCoverArtHandler(
	mediaStorage *storage.MediaStorage,
) *getCoverArtHandler {
	return &getCoverArtHandler{
		mediaStorage: mediaStorage,
	}
}

func (h *getCoverArtHandler) Handle(w http.ResponseWriter, r *http.Request) (*responses.SubsonicResponse, error) {
	rawId := r.URL.Query().Get("id")
	if rawId == "" {
		return responses.NewParameterMissingResponse("id"), nil
	}

	var cover storage.CoverDescriptor
	var err error
	if strings.HasPrefix(rawId, "playlist/") {
		id, e := uuid.Parse(strings.TrimPrefix(rawId, "playlist/"))
		if e != nil {
			err = e
		} else {
			cover, err = h.mediaStorage.GetPlaylistCover(id)
		}
	} else if strings.HasPrefix(rawId, "album/") {
		id, e := uuid.Parse(strings.TrimPrefix(rawId, "album/"))
		if e != nil {
			err = e
		} else {
			cover, err = h.mediaStorage.GetAlbumCover(id)
		}
	} else {
		return responses.NewNotFoundResponse(fmt.Sprintf("Cover art `%s`", rawId)), nil
	}

	if err != nil {
		return nil, err
	}

	reader, err := os.Open(cover.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	w.Header().Add("Content-Type", util.FormatToMediaType(cover.Format))
	_, err = io.Copy(w, reader)
	return nil, err
}
