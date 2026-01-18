package subsonic

import (
	"net/http"

	"tapesonic/library"
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/util"
)

func Search3(auth *authenticator, librarySvc *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		query := r.URL.Query().Get("query")

		albumCount := util.StringToIntOrDefault(r.URL.Query().Get("albumCount"), 20)
		albumOffset := util.StringToIntOrDefault(r.URL.Query().Get("albumOffset"), 0)
		var albums []model.LibraryAlbum
		if query == "" {
			albums, err = librarySvc.GetAlbumsSortById(user, albumCount, albumOffset)
		} else {
			albums, err = librarySvc.SearchAlbumsByQuery(user, query, albumCount, albumOffset)
		}
		if err != nil {
			return nil, err
		}

		artistCount := util.StringToIntOrDefault(r.URL.Query().Get("artistCount"), 20)
		artistOffset := util.StringToIntOrDefault(r.URL.Query().Get("artistOffset"), 0)
		var artists []model.LibraryArtist
		if query == "" {
			artists, err = librarySvc.GetArtistsSortById(user, artistCount, artistOffset)
		} else {
			artists, err = librarySvc.SearchArtistsByQuery(user, query, artistCount, artistOffset)
		}
		if err != nil {
			return nil, err
		}

		songCount := util.StringToIntOrDefault(r.URL.Query().Get("songCount"), 20)
		songOffset := util.StringToIntOrDefault(r.URL.Query().Get("songOffset"), 0)
		var tracks []model.LibraryTrack
		if query == "" {
			tracks, err = librarySvc.GetTracksSortById(user, songCount, songOffset)
		} else {
			tracks, err = librarySvc.SearchTracksByFields(user.Id, library.TrackFilter{Title: query}, songCount, songOffset)
		}
		if err != nil {
			return nil, err
		}

		response := subsonic.NewOkResponse()
		response.SearchResult3 = &subsonic.SearchResult3{
			Artist: util.Map(artists, ToArtistId3),
			Album:  util.Map(albums, ToAlbumId3),
			Song:   util.Map(tracks, ToChild),
		}
		return response, nil
	}
}
