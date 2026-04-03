package subsonic

import (
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/util"
	"time"
)

func ToPlaylist(playlist model.LibraryPlaylist) subsonic.Playlist {
	rs := subsonic.Playlist{
		Id:        playlist.Id,
		Name:      playlist.Name,
		SongCount: playlist.TrackCount,
		Duration:  int(playlist.Duration.Seconds()),
		Created:   playlist.CreatedAt.Unwrap(),
		Changed:   playlist.UpdatedAt.Unwrap(),
		// Owner
		Entry: util.Map(playlist.Tracks, ToChild),
	}

	if playlist.CoverId != nil {
		rs.CoverArt = playlist.CoverId.String()
	}

	return rs
}

func ToAlbumId3(album model.LibraryAlbum) subsonic.AlbumId3 {
	var year = 0
	var releaseDate *subsonic.ItemDate = nil
	if album.ReleasedAt != nil {
		year = album.ReleasedAt.Unwrap().Year()
		date := ToItemDate(album.ReleasedAt.Unwrap())
		releaseDate = &date
	}

	rs := subsonic.AlbumId3{
		Id:        album.Id,
		Name:      album.Name,
		Artist:    album.ArtistName,
		ArtistId:  album.ArtistId,
		SongCount: album.TrackCount,
		Duration:  int(album.Duration.Seconds()),
		PlayCount: 0,
		Created:   album.AddedAt.Unwrap(),
		// Starred:     album.StarredAt,
		Played:      album.PlayedAt.UnwrapNullable(),
		Year:        year,
		ReleaseDate: releaseDate,
		Song:        util.Map(album.Tracks, ToChild),
	}

	if album.CoverId != nil {
		rs.CoverArt = album.CoverId.String()
	}

	return rs
}

func ToArtistId3(artist model.LibraryArtist) subsonic.ArtistId3 {
	rs := subsonic.ArtistId3{
		Id:         artist.Id,
		Name:       artist.Name,
		AlbumCount: artist.AlbumCount,
		Album:      util.Map(artist.Albums, ToAlbumId3),
	}

	if artist.CoverId != nil {
		rs.CoverArt = artist.CoverId.String()
	}

	return rs
}

func ToItemDate(date time.Time) subsonic.ItemDate {
	return subsonic.ItemDate{
		Year:  date.Year(),
		Month: int(date.Month()),
		Day:   date.Day(),
	}
}

func ToChild(track model.LibraryTrack) subsonic.Child {
	result := subsonic.Child{
		Id:        track.Id,
		IsDir:     false,
		Artist:    track.ArtistName,
		Title:     track.Title,
		Album:     track.AlbumName,
		AlbumId:   track.AlbumId,
		Track:     track.AlbumTrackIndex + 1,
		Duration:  int(track.Duration.Seconds()),
		PlayCount: 0,
		Played:    track.PlayedAt.UnwrapNullable(),
	}

	if track.ArtistId != nil {
		result.ArtistId = track.ArtistId.String()
	}
	if track.CoverId != nil {
		result.CoverArt = track.CoverId.String()
	}

	return result
}
