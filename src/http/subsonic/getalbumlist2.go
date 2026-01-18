package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/util"
)

func GetAlbumList2(auth *authenticator, library *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		listType := r.URL.Query().Get("type")
		if listType == "" {
			return subsonic.NewParameterMissingResponse("type"), nil
		}

		size := util.StringToIntOrDefault(r.URL.Query().Get("size"), 10)
		offset := util.StringToIntOrDefault(r.URL.Query().Get("offset"), 0)
		fromYear := util.StringToIntOrNull(r.URL.Query().Get("fromYear"))
		toYear := util.StringToIntOrNull(r.URL.Query().Get("toYear"))

		var albums []model.LibraryAlbum
		switch listType {
		case subsonic.ALBUM_LIST_RANDOM:
			albums, err = library.GetRandomAlbums(user, size)
		case subsonic.ALBUM_LIST_NEWEST:
			albums, err = library.GetAlbumsSortByAddedAtDesc(user, size, offset)
		case subsonic.ALBUM_LIST_HIGHEST:
			break // todo
		case subsonic.ALBUM_LIST_FREQUENT:
			albums, err = library.GetAlbumsSortByTotalListenTimeDesc(user, size, offset)
		case subsonic.ALBUM_LIST_RECENT:
			albums, err = library.GetAlbumsSortByPlayedAtDesc(user, size, offset)
		case subsonic.ALBUM_LIST_BY_NAME:
			albums, err = library.GetAlbumsSortByTitle(user, size, offset)
		case subsonic.ALBUM_LIST_BY_ARTIST:
			albums, err = library.GetAlbumsSortByArtist(user, size, offset)
		case subsonic.ALBUM_LIST_STARRED:
			break // todo
		case subsonic.ALBUM_LIST_BY_YEAR:
			if fromYear == nil {
				return subsonic.NewParameterMissingResponse("fromYear"), nil
			}
			if toYear == nil {
				return subsonic.NewParameterMissingResponse("toYear"), nil
			}

			albums, err = library.GetAlbumsSortByReleasedAtDesc(user, size, offset, *fromYear, *toYear)
		case subsonic.ALBUM_LIST_BY_GENRE:
			break // todo
		}

		if err != nil {
			return nil, err
		}

		response := subsonic.NewOkResponse()
		response.AlbumList2 = &subsonic.AlbumList2{
			Album: util.Map(albums, ToAlbumId3),
		}
		return response, nil
	}
}
