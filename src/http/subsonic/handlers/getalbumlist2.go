package handlers

import (
	"fmt"
	"net/http"

	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
	"tapesonic/util"
)

type getAlbumList2Handler struct {
	albumStorage *storage.AlbumStorage
}

func NewGetAlbumList2Handler(albumStorage *storage.AlbumStorage) *getAlbumList2Handler {
	return &getAlbumList2Handler{
		albumStorage: albumStorage,
	}
}

const (
	LIST_RANDOM = "random"
	LIST_NEWEST = "newest"
	// LIST_HIGHEST   = "highest" // todo
	// LIST_FREQUENT  = "frequent" // todo
	// LIST_RECENT    = "recent" // todo
	LIST_BY_NAME   = "alphabeticalByName"
	LIST_BY_ARTIST = "alphabeticalByArtist"
)

func (h *getAlbumList2Handler) Handle(r *http.Request) (*responses.SubsonicResponse, error) {
	listType := r.URL.Query().Get("type")
	if listType == "" {
		return responses.NewParameterMissingResponse("type"), nil
	}

	size := util.StringToIntOrDefault(r.URL.Query().Get("size"), 10)
	offset := util.StringToIntOrDefault(r.URL.Query().Get("offset"), 0)

	var albums []storage.SubsonicAlbumListItem
	var err error
	if listType == LIST_RANDOM {
		albums, err = h.albumStorage.GetSubsonicAlbumsSortRandom(size, offset)
	} else if listType == LIST_NEWEST {
		albums, err = h.albumStorage.GetSubsonicAlbumsSortNewest(size, offset)
	} else if listType == LIST_BY_NAME {
		albums, err = h.albumStorage.GetSubsonicAlbumsSortName(size, offset)
	} else if listType == LIST_BY_ARTIST {
		albums, err = h.albumStorage.GetSubsonicAlbumsSortArtist(size, offset)
	} else {
		return responses.NewServerErrorResponse(fmt.Sprintf("Unsupported album sort order %s", listType)), nil
	}

	if err != nil {
		return nil, err
	}

	albumsResponse := []responses.AlbumId3{}
	for _, album := range albums {
		albumResponse := responses.NewAlbumId3(
			fmt.Sprint(album.Album.Id),
			album.Album.Name,
			album.Album.Artist,
			"album/"+fmt.Sprint(album.Album.Id),
			album.SongCount,
			album.DurationSec,
			album.Album.CreatedAt,
		)
		albumsResponse = append(albumsResponse, *albumResponse)
	}

	response := responses.NewOkResponse()
	response.AlbumList2 = responses.NewAlbumList2(albumsResponse)
	return response, nil
}
