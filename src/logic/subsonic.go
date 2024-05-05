package logic

import (
	"context"
	"io"
	"tapesonic/http/subsonic/responses"
)

type SubsonicService interface {
	GetAlbum(id string) (*responses.AlbumId3, error)

	GetAlbumList2(
		type_ string,
		size int,
		offset int,
	) (*responses.AlbumList2, error)

	GetPlaylist(id string) (*responses.SubsonicPlaylist, error)

	GetPlaylists() (*responses.SubsonicPlaylists, error)

	GetCoverArt(id string) (mime string, reader io.ReadCloser, err error)

	Stream(ctx context.Context, id string) (mime string, reader io.ReadCloser, err error)
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
