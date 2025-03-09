package logic

import (
	"context"
	"io"
	"tapesonic/http/subsonic/responses"
	"time"
)

type SubsonicService interface {
	Search3(
		query string,
		artistCount int,
		artistOffset int,
		albumCount int,
		albumOffset int,
		songCount int,
		songOffset int,
	) (*responses.SearchResult3, error)

	GetSong(id string) (*responses.SubsonicChild, error)

	GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error)

	GetAlbum(id string) (*responses.AlbumId3, error)

	GetAlbumList2(
		type_ string,
		size int,
		offset int,
		fromYear *int,
		toYear *int,
	) (*responses.AlbumList2, error)

	GetPlaylist(id string) (*responses.SubsonicPlaylist, error)

	GetPlaylists() (*responses.SubsonicPlaylists, error)

	GetArtist(id string) (*responses.Artist, error)

	Scrobble(id string, time_ time.Time, submission bool) error

	GetCoverArt(id string) (mime string, reader io.ReadCloser, err error)

	Stream(ctx context.Context, id string) (AudioStream, error)

	GetLicense() (*responses.License, error)
}

const (
	LIST_RANDOM = "random"
	LIST_NEWEST = "newest"
	// LIST_HIGHEST   = "highest" // todo
	LIST_FREQUENT  = "frequent"
	LIST_RECENT    = "recent"
	LIST_BY_NAME   = "alphabeticalByName"
	LIST_BY_ARTIST = "alphabeticalByArtist"
	LIST_STARRED   = "starred"
	LIST_BY_YEAR   = "byYear"
)
