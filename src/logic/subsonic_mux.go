package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"sort"
	"strings"
	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
	"tapesonic/util"
	"time"
)

const (
	albumListFetchSize = 500
)

type SubsonicMuxService struct {
	services map[string]SubsonicService

	cachedMuxSong    *storage.CachedMuxSongStorage
	muxedSongListens *storage.MuxedSongListensStorage
}

func NewSubsonicMuxService(
	cachedMuxSong *storage.CachedMuxSongStorage,
	muxedSongListens *storage.MuxedSongListensStorage,
) *SubsonicMuxService {
	return &SubsonicMuxService{
		services:         map[string]SubsonicService{},
		cachedMuxSong:    cachedMuxSong,
		muxedSongListens: muxedSongListens,
	}
}

func (svc *SubsonicMuxService) AddService(prefix string, service SubsonicService) {
	svc.services[prefix] = service
}

func (svc *SubsonicMuxService) GetSong(prefixedId string) (*responses.SubsonicChild, error) {
	serviceName, service, err := svc.findService(prefixedId)
	if err != nil {
		return nil, err
	}

	song, err := service.GetSong(removePrefix(serviceName, prefixedId))
	if err != nil {
		return nil, err
	}

	_, cacheWriteErr := svc.cachedMuxSong.Save(serviceName, song.Id, song.AlbumId, song.Duration)
	if cacheWriteErr != nil {
		slog.Error(fmt.Sprintf("Failed to cache song info when proxying getSong: %s", cacheWriteErr.Error()))
	}

	song.Id = addPrefix(serviceName, song.Id)
	song.CoverArt = addPrefix(serviceName, song.CoverArt)
	song.AlbumId = addPrefix(serviceName, song.AlbumId)

	return song, nil
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

	rewrittenAlbum := rewriteAlbumInfo(serviceName, *album)
	album = &rewrittenAlbum
	for i := range album.Song {
		song := &album.Song[i]
		song.Id = addPrefix(serviceName, song.Id)
		song.CoverArt = addPrefix(serviceName, song.CoverArt)
		song.AlbumId = addPrefix(serviceName, song.AlbumId)
	}

	return album, nil
}

func (svc *SubsonicMuxService) GetAlbumList2(
	type_ string,
	size int,
	offset int,
	fromYear *int,
	toYear *int,
) (*responses.AlbumList2, error) {
	if len(svc.services) == 1 {
		for serviceName, service := range svc.services {
			albums, err := service.GetAlbumList2(type_, size, offset, fromYear, toYear)
			if err != nil {
				return nil, err
			}

			for i := range albums.Album {
				albums.Album[i] = rewriteAlbumInfo(serviceName, albums.Album[i])
			}

			return albums, nil
		}
	}

	if type_ == LIST_RECENT || type_ == LIST_FREQUENT {
		var albumListenStats []storage.MuxedAlbumListenStats
		var err error
		if type_ == LIST_RECENT {
			albumListenStats, err = svc.muxedSongListens.GetRecentAlbumListenStats(size, offset)
		} else {
			albumListenStats, err = svc.muxedSongListens.GetFrequentAlbumListenStats(size, offset)
		}
		if err != nil {
			return nil, err
		}

		albums, err := util.ParallelMap(albumListenStats, func(item storage.MuxedAlbumListenStats) (responses.AlbumId3, error) {
			service, err := svc.findServiceByName(item.ServiceName)
			if err != nil {
				return responses.AlbumId3{}, err
			}

			album, err := service.GetAlbum(item.Id)
			if err != nil {
				return responses.AlbumId3{}, err
			}

			album.Song = []responses.SubsonicChild{}
			return rewriteAlbumInfo(item.ServiceName, *album), nil
		})
		if err != nil {
			return nil, err
		}

		return responses.NewAlbumList2(albums), nil
	}

	albums := []responses.AlbumId3{}
	for serviceName, service := range svc.services {
		// yes, this is absolutely disgusting;
		// but it's the only way to keep the sorting/pagination stable between different backing servers
		// and also work around the fact that some servers don't follow the specification;
		// will be solved properly later by just caching the complete album list in the database
		serviceOffset := 0
		for {
			more, err := service.GetAlbumList2(type_, albumListFetchSize, serviceOffset, fromYear, toYear)
			if err != nil {
				return nil, err
			}

			for i := range more.Album {
				more.Album[i] = rewriteAlbumInfo(serviceName, more.Album[i])
			}

			albums = append(albums, more.Album...)

			if len(more.Album) < albumListFetchSize {
				break
			} else {
				serviceOffset += len(more.Album)
			}
		}
	}

	switch type_ {
	case LIST_RANDOM:
		rand.Shuffle(len(albums), func(i int, j int) { albums[i], albums[j] = albums[j], albums[i] })
	case LIST_NEWEST:
		sort.Slice(albums, func(i, j int) bool {
			return albums[i].Created.After(albums[j].Created)
		})
	case LIST_BY_NAME:
		sort.Slice(albums, func(i, j int) bool {
			return strings.ToLower(albums[i].Name) < strings.ToLower(albums[j].Name)
		})
	case LIST_BY_ARTIST:
		sort.Slice(albums, func(i, j int) bool {
			return strings.ToLower(albums[i].Artist) < strings.ToLower(albums[j].Artist)
		})
	case LIST_STARRED:
		sort.Slice(albums, func(i, j int) bool {
			// todo: filter no-starred-date albums out; pushing those to the end for now
			if albums[i].Starred == nil {
				return false
			}
			if albums[j].Starred == nil {
				return true
			}

			if albums[i].Starred != albums[j].Starred {
				return (*albums[i].Starred).After(*albums[j].Starred)
			} else {
				return albums[i].Created.After(albums[j].Created)
			}
		})
	case LIST_BY_YEAR:
		sort.Slice(albums, func(i, j int) bool {
			// todo: filter no-release-date albums out; pushing those to the end for now
			if albums[i].Year == 0 {
				return false
			}
			if albums[j].Year == 0 {
				return true
			}

			if *fromYear > *toYear {
				i, j = j, i
			}

			if albums[i].Year != albums[j].Year {
				return albums[i].Year < albums[j].Year
			} else {
				return albums[i].Created.Before(albums[j].Created)
			}
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
		entry.CoverArt = addPrefix(serviceName, entry.CoverArt)
		entry.AlbumId = addPrefix(serviceName, entry.AlbumId)
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

	unprefixedId := removePrefix(serviceName, prefixedId)

	song, cacheWriteErr := service.GetSong(unprefixedId)
	if cacheWriteErr == nil {
		_, cacheWriteErr = svc.cachedMuxSong.Save(serviceName, song.Id, song.AlbumId, song.Duration)
	}
	if cacheWriteErr != nil {
		slog.Error(fmt.Sprintf("Failed to cache song info when scrobbling: %s", cacheWriteErr.Error()))
	}

	selfErr := svc.muxedSongListens.Record(serviceName, unprefixedId, time_, submission)
	serviceErr := service.Scrobble(unprefixedId, time_, submission)

	return errors.Join(selfErr, serviceErr)
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

func rewriteAlbumInfo(serviceName string, album responses.AlbumId3) responses.AlbumId3 {
	album.Id = addPrefix(serviceName, album.Id)
	album.CoverArt = addPrefix(serviceName, album.CoverArt)
	return album
}

func (svc *SubsonicMuxService) findServiceByName(serviceName string) (SubsonicService, error) {
	service := svc.services[serviceName]
	if service == nil {
		return nil, fmt.Errorf("unknown subsonic service `%s`", serviceName)
	} else {
		return service, nil
	}
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
	if unprefixedId == "" {
		return ""
	}

	return generatePrefix(serviceName) + unprefixedId
}

func removePrefix(serviceName string, prefixedId string) string {
	prefix := generatePrefix(serviceName)
	return strings.TrimPrefix(prefixedId, prefix)
}

func generatePrefix(serviceName string) string {
	return "@" + serviceName + "/"
}
