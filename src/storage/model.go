package storage

import (
	"time"

	"github.com/google/uuid"
)

type RelatedItems struct {
	Tapes     []*Tape
	Playlists []*Playlist
	Albums    []*Album
}

type SubsonicAlbumItem struct {
	Id uuid.UUID

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
	Id uuid.UUID

	Name   string
	Artist string

	ThumbnailPath string

	CreatedAt time.Time
	UpdatedAt time.Time

	SongCount   int
	DurationSec int
}

type SubsonicTrackItem struct {
	Id uuid.UUID

	AlbumId uuid.UUID

	AlbumTrackIndex    int
	PlaylistTrackIndex int

	Album  string
	Artist string
	Title  string

	DurationSec int
	PlayCount   int
}

type MuxedAlbumListenStats struct {
	ServiceName string
	Id          string
}
