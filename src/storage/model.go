package storage

import (
	"time"

	"github.com/google/uuid"
)

type TrackForTapeMetadataGuessing struct {
	Id uuid.UUID

	AlbumArtist        string
	AlbumTitle         string
	Artist             string
	SourceTitle        string
	SourceParentTitles []string `gorm:"serializer:json"`

	ReleaseDate *time.Time
	ThumbnailId *uuid.UUID
}

type SourceForHierarchy struct {
	Id       uuid.UUID
	ParentId *uuid.UUID

	Url      string
	Title    string
	Uploader string

	ListIndex int

	ThumbnailId *uuid.UUID
}

type SubsonicAlbumItem struct {
	Id string

	Name        string
	Artist      string
	ReleaseDate *time.Time

	ThumbnailId *uuid.UUID

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

	ThumbnailId *uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time

	SongCount   int
	DurationSec int
}

type SubsonicTrackItem struct {
	Id          string
	ThumbnailId *uuid.UUID

	AlbumId          string
	AlbumThumbnailId *uuid.UUID

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
