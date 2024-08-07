package storage

import (
	"time"
)

type RelatedItems struct {
	Tapes     []*Tape
	Playlists []*Playlist
	Albums    []*Album
}

type SubsonicAlbumItem struct {
	Id string

	Name        string
	Artist      string
	ReleaseDate *time.Time

	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	SongCount   int
	DurationSec int
	PlayCount   int
}

type SubsonicPlaylistItem struct {
	Id string

	CreatedBy string

	Name   string
	Artist string

	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	SongCount   int
	DurationSec int
}

type SubsonicTrackItem struct {
	Id string

	AlbumId string

	AlbumTrackIndex    int
	PlaylistTrackIndex int

	Album  string
	Artist string
	Title  string

	DurationSec int
	PlayCount   int
}

type CachedArtistId struct {
	ServiceName string
	Id          string
}

type CachedAlbumId struct {
	ServiceName string
	Id          string
}

type CachedSongId struct {
	ServiceName string
	Id          string
}

type CachedSongIdWithIndex struct {
	CachedSongId

	TrackIndex int
}
