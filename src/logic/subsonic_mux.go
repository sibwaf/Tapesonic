package logic

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"tapesonic/http/subsonic/responses"
	"time"
)

type SubsonicMuxService struct {
	services map[string]SubsonicService
}

func NewSubsonicMuxService() *SubsonicMuxService {
	return &SubsonicMuxService{
		services: map[string]SubsonicService{},
	}
}

func (svc *SubsonicMuxService) AddService(prefix string, service SubsonicService) {
	svc.services[prefix] = service
}

func (svc *SubsonicMuxService) GetAlbum(prefixedId string) (*responses.AlbumId3, error) {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return nil, err
	}

	album, err := service.GetAlbum(removePrefix(serviceName, prefixedId))
	if err != nil {
		return nil, err
	}

	album.Id = addPrefix(serviceName, album.Id)
	album.CoverArt = addPrefix(serviceName, album.CoverArt)
	for i := range album.Song {
		song := &album.Song[i]
		song.Id = addPrefix(serviceName, song.Id)
	}

	return album, nil
}

func (svc *SubsonicMuxService) GetAlbumList2(
	type_ string,
	size int,
	offset int,
) (*responses.AlbumList2, error) {
	if len(svc.services) == 1 {
		for serviceName, service := range svc.services {
			albums, err := service.GetAlbumList2(type_, size, offset)
			if err != nil {
				return nil, err
			}

			for i := range albums.Album {
				album := &albums.Album[i]
				album.Id = addPrefix(serviceName, album.Id)
				album.CoverArt = addPrefix(serviceName, album.CoverArt)
			}

			return albums, nil
		}
	}

	albums := []responses.AlbumId3{}
	for serviceName, service := range svc.services {
		more, err := service.GetAlbumList2(type_, offset+size, 0)
		if err != nil {
			return nil, err
		}

		for i := range more.Album {
			album := &more.Album[i]
			album.Id = addPrefix(serviceName, album.Id)
			album.CoverArt = addPrefix(serviceName, album.CoverArt)
		}

		albums = append(albums, more.Album...)
	}

	switch type_ {
	case LIST_RANDOM:
		rand.Shuffle(len(albums), func(i int, j int) { albums[i], albums[j] = albums[j], albums[i] })
	case LIST_NEWEST:
		sort.Slice(albums, func(i, j int) bool {
			created1, err := time.Parse(time.RFC3339, albums[i].Created)
			if err != nil {
				return false
			}

			created2, err := time.Parse(time.RFC3339, albums[j].Created)
			if err != nil {
				return true
			}

			return created1.After(created2)
		})
	case LIST_BY_NAME:
		sort.Slice(albums, func(i, j int) bool {
			return strings.ToLower(albums[i].Name) < strings.ToLower(albums[j].Name)
		})
	case LIST_BY_ARTIST:
		sort.Slice(albums, func(i, j int) bool {
			return strings.ToLower(albums[i].Artist) < strings.ToLower(albums[j].Artist)
		})
	default:
		return nil, fmt.Errorf("unsupported type=%s in getAlbumList2", type_)
	}

	var listEnd int
	if len(albums) < offset+size {
		listEnd = len(albums)
	} else {
		listEnd = offset + size
	}
	albums = albums[offset:listEnd]

	return responses.NewAlbumList2(albums), nil
}

func (svc *SubsonicMuxService) GetPlaylist(prefixedId string) (*responses.SubsonicPlaylist, error) {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return nil, err
	}

	playlist, err := service.GetPlaylist(removePrefix(serviceName, prefixedId))
	if err != nil {
		return nil, err
	}

	playlist.Id = addPrefix(serviceName, playlist.Id)
	playlist.CoverArt = addPrefix(serviceName, playlist.CoverArt)
	for i := range playlist.Entry {
		entry := &playlist.Entry[i]
		entry.Id = addPrefix(serviceName, entry.Id)
	}

	return playlist, nil
}

func (svc *SubsonicMuxService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists := []responses.SubsonicPlaylist{}

	for serviceName, service := range svc.services {
		servicePlaylists, err := service.GetPlaylists()
		if err != nil {
			return nil, err
		}

		for i := range servicePlaylists.Playlist {
			playlist := &servicePlaylists.Playlist[i]
			playlist.Id = addPrefix(serviceName, playlist.Id)
			playlist.CoverArt = addPrefix(serviceName, playlist.CoverArt)
		}

		playlists = append(playlists, servicePlaylists.Playlist...)
	}

	return responses.NewSubsonicPlaylists(playlists), nil
}

func (svc *SubsonicMuxService) Scrobble(prefixedId string, time_ time.Time, submission bool) error {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return err
	}

	return service.Scrobble(removePrefix(serviceName, prefixedId), time_, submission)
}

func (svc *SubsonicMuxService) GetCoverArt(prefixedId string) (mime string, reader io.ReadCloser, err error) {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return
	}

	return service.GetCoverArt(removePrefix(serviceName, prefixedId))
}

func (svc *SubsonicMuxService) Stream(ctx context.Context, prefixedId string) (mime string, reader io.ReadCloser, err error) {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return
	}

	return service.Stream(ctx, removePrefix(serviceName, prefixedId))
}

func (svc *SubsonicMuxService) findService(prefixedId string) (string, SubsonicService, error) {
	for name, service := range svc.services {
		prefix := generatePrefix(name)
		if strings.HasPrefix(prefixedId, prefix) {
			return name, service, nil
		}
	}

	return "", nil, fmt.Errorf("failed to find the backing subsonic service for id `%s`", prefixedId)
}

func addPrefix(serviceName string, unprefixedId string) string {
	return generatePrefix(serviceName) + unprefixedId
}

func removePrefix(serviceName string, prefixedId string) string {
	prefix := generatePrefix(serviceName)
	return strings.TrimPrefix(prefixedId, prefix)
}

func generatePrefix(serviceName string) string {
	return "@" + serviceName + "/"
}
