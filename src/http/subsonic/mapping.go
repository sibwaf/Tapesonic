package subsonic

import (
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/util"
	"time"
)

func ToPlaylist(playlist model.LibraryPlaylist) subsonic.Playlist {
	return subsonic.Playlist{
		Id:        playlist.Id,
		Name:      playlist.Name,
		SongCount: playlist.TrackCount,
		Duration:  int(playlist.Duration.Seconds()),
		Created:   playlist.CreatedAt.Unwrap(),
		Changed:   playlist.UpdatedAt.Unwrap(),
		CoverArt:  playlist.CoverId,
		// Owner
		Entry: util.Map(playlist.Tracks, ToChild),
	}
}

func ToAlbumId3(album model.LibraryAlbum) subsonic.AlbumId3 {
	var year = 0
	var releaseDate *subsonic.ItemDate = nil
	if album.ReleasedAt != nil {
		year = album.ReleasedAt.Unwrap().Year()
		date := ToItemDate(album.ReleasedAt.Unwrap())
		releaseDate = &date
	}

	return subsonic.AlbumId3{
		Id:        album.Id,
		Name:      album.Name,
		Artist:    album.ArtistName,
		ArtistId:  album.ArtistId,
		CoverArt:  album.CoverId,
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
}

func ToArtistId3(artist model.LibraryArtist) subsonic.ArtistId3 {
	return subsonic.ArtistId3{
		Id:         artist.Id,
		Name:       artist.Name,
		CoverArt:   artist.CoverId,
		AlbumCount: artist.AlbumCount,
		Album:      util.Map(artist.Albums, ToAlbumId3),
	}
}

func ToItemDate(date time.Time) subsonic.ItemDate {
	return subsonic.ItemDate{
		Year:  date.Year(),
		Month: int(date.Month()),
		Day:   date.Day(),
	}
}

func ToChild(track model.LibraryTrack) subsonic.Child {
	return subsonic.Child{
		Id:        track.Id,
		IsDir:     false,
		Artist:    track.ArtistName,
		ArtistId:  track.ArtistId,
		Title:     track.Title,
		Album:     track.AlbumName,
		AlbumId:   track.AlbumId,
		Track:     track.AlbumTrackIndex,
		CoverArt:  track.CoverId,
		Duration:  int(track.Duration.Seconds()),
		PlayCount: 0,
		Played:    track.PlayedAt.UnwrapNullable(),
	}
}
