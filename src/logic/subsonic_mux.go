package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
	"tapesonic/util"
	"time"
)

const (
	fetchSize = 500
)

type SubsonicMuxService struct {
	services []*SubsonicNamedService

	cachedMuxSong    *storage.CachedMuxSongStorage
	muxedSongListens *storage.MuxedSongListensStorage

	scrobbler *ScrobbleService
}

func NewSubsonicMuxService(
	cachedMuxSong *storage.CachedMuxSongStorage,
	muxedSongListens *storage.MuxedSongListensStorage,
	scrobbler *ScrobbleService,
) *SubsonicMuxService {
	return &SubsonicMuxService{
		services:         []*SubsonicNamedService{},
		cachedMuxSong:    cachedMuxSong,
		muxedSongListens: muxedSongListens,
		scrobbler:        scrobbler,
	}
}

func (svc *SubsonicMuxService) AddService(service *SubsonicNamedService) {
	svc.services = append(svc.services, service)
}

func (svc *SubsonicMuxService) Search3(
	query string,
	artistCount int,
	artistOffset int,
	albumCount int,
	albumOffset int,
	songCount int,
	songOffset int,
) (*responses.SearchResult3, error) {
	if len(svc.services) == 1 {
		for _, service := range svc.services {
			return service.Search3(query, artistCount, artistOffset, albumCount, albumOffset, songCount, songOffset)
		}
	}

	artists := []responses.ArtistId3{}
	albums := []responses.AlbumId3{}
	songs := []responses.SubsonicChild{}

	for _, service := range svc.services {
		// i hate this; see more in getAlbumList2
		serviceArtistOffset := 0
		serviceAlbumOffset := 0
		serviceSongOffset := 0
		for {
			searchResult, err := service.Search3(query, fetchSize, serviceArtistOffset, fetchSize, serviceAlbumOffset, fetchSize, serviceSongOffset)
			if err != nil {
				return nil, err
			}

			serviceArtistOffset += len(searchResult.Artist)
			artists = append(artists, searchResult.Artist...)

			serviceAlbumOffset += len(searchResult.Album)
			albums = append(albums, searchResult.Album...)

			serviceSongOffset += len(searchResult.Song)
			songs = append(songs, searchResult.Song...)

			if len(searchResult.Artist) < fetchSize && len(searchResult.Album) < fetchSize && len(searchResult.Song) < fetchSize {
				break
			}
		}
	}

	slices.SortFunc(artists, func(a responses.ArtistId3, b responses.ArtistId3) int { return strings.Compare(a.Id, b.Id) })
	slices.SortFunc(albums, func(a responses.AlbumId3, b responses.AlbumId3) int { return strings.Compare(a.Id, b.Id) })
	slices.SortFunc(songs, func(a responses.SubsonicChild, b responses.SubsonicChild) int { return strings.Compare(a.Id, b.Id) })

	return &responses.SearchResult3{
		Artist: artists[min(artistOffset, len(artists)):min(artistOffset+artistCount, len(artists))],
		Album:  albums[min(albumOffset, len(albums)):min(albumOffset+albumCount, len(albums))],
		Song:   songs[min(songOffset, len(songs)):min(songOffset+songCount, len(songs))],
	}, nil
}

func (svc *SubsonicMuxService) GetSong(id string) (*responses.SubsonicChild, error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return nil, err
	}

	return service.GetSong(id)
}

func (svc *SubsonicMuxService) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	songs := []responses.SubsonicChild{}
	for _, service := range svc.services {
		// todo: a pretty bad implementation, but it makes at least a somewhat more balanced result
		// when different services have a different count of total songs than just getting `size` songs from each one
		more, err := service.GetRandomSongs(fetchSize, genre, fromYear, toYear)
		if err != nil {
			return nil, err
		}

		songs = append(songs, more.Song...)
	}

	rand.Shuffle(len(songs), func(i int, j int) { songs[i], songs[j] = songs[j], songs[i] })

	songs = songs[:min(size, len(songs))]

	return responses.NewRandomSongs(songs), nil
}

func (svc *SubsonicMuxService) GetAlbum(id string) (*responses.AlbumId3, error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return nil, err
	}

	return service.GetAlbum(id)
}

func (svc *SubsonicMuxService) GetAlbumList2(
	type_ string,
	size int,
	offset int,
	fromYear *int,
	toYear *int,
) (*responses.AlbumList2, error) {
	if len(svc.services) == 1 {
		for _, service := range svc.services {
			return service.GetAlbumList2(type_, size, offset, fromYear, toYear)
		}
	}

	if type_ == LIST_RECENT || type_ == LIST_FREQUENT {
		var albumListenStats []storage.CachedAlbumId
		var err error
		if type_ == LIST_RECENT {
			albumListenStats, err = svc.muxedSongListens.GetRecentAlbumListenStats(size, offset)
		} else {
			albumListenStats, err = svc.muxedSongListens.GetFrequentAlbumListenStats(size, offset)
		}
		if err != nil {
			return nil, err
		}

		albums, err := util.ParallelMap(albumListenStats, func(item storage.CachedAlbumId) (responses.AlbumId3, error) {
			service, err := svc.findServiceByName(item.ServiceName)
			if err != nil {
				return responses.AlbumId3{}, err
			}

			album, err := service.GetAlbumByRawId(item.Id)
			if err != nil {
				return responses.AlbumId3{}, err
			}

			album.Song = []responses.SubsonicChild{}
			return *album, nil
		})
		if err != nil {
			return nil, err
		}

		return responses.NewAlbumList2(albums), nil
	}

	albums := []responses.AlbumId3{}
	for _, service := range svc.services {
		// yes, this is absolutely disgusting;
		// but it's the only way to keep the sorting/pagination stable between different backing servers
		// and also work around the fact that some servers don't follow the specification;
		// will be solved properly later by just caching the complete album list in the database
		serviceOffset := 0
		for {
			more, err := service.GetAlbumList2(type_, fetchSize, serviceOffset, fromYear, toYear)
			if err != nil {
				return nil, err
			}

			albums = append(albums, more.Album...)

			if len(more.Album) < fetchSize {
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

	albums = albums[min(offset, len(albums)):min(offset+size, len(albums))]

	return responses.NewAlbumList2(albums), nil
}

func (svc *SubsonicMuxService) GetPlaylist(id string) (*responses.SubsonicPlaylist, error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return nil, err
	}

	return service.GetPlaylist(id)
}

func (svc *SubsonicMuxService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists := []responses.SubsonicPlaylist{}

	for _, service := range svc.services {
		servicePlaylists, err := service.GetPlaylists()
		if err != nil {
			return nil, err
		}

		playlists = append(playlists, servicePlaylists.Playlist...)
	}

	return responses.NewSubsonicPlaylists(playlists), nil
}

func (svc *SubsonicMuxService) GetArtist(id string) (*responses.Artist, error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return nil, err
	}

	return service.GetArtist(id)
}

func (svc *SubsonicMuxService) Scrobble(id string, time_ time.Time, submission bool) error {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return err
	}

	song, cacheWriteErr := service.GetSong(id)
	if cacheWriteErr == nil {
		rawSong := service.GetRawSong(*song)
		_, cacheWriteErr = svc.cachedMuxSong.Save(
			storage.CachedMuxSong{
				ServiceName: service.Name(),
				SongId:      rawSong.Id,
				AlbumId:     rawSong.AlbumId,
				Artist:      rawSong.Artist,
				Album:       rawSong.Album,
				Title:       rawSong.Title,
				DurationSec: rawSong.Duration,
				CachedAt:    time.Now(),
			},
		)
	}
	if cacheWriteErr != nil {
		slog.Error(fmt.Sprintf("Failed to cache song info when scrobbling: %s", cacheWriteErr.Error()))
	}

	selfErr := svc.muxedSongListens.Record(service.Name(), service.RemovePrefix(id), time_, submission)
	serviceErr := service.Scrobble(id, time_, submission)
	scrobblerErr := svc.scrobbleWithScrobbler(service.Name(), service.RemovePrefix(id), time_, submission)

	return errors.Join(selfErr, serviceErr, scrobblerErr)
}

func (svc *SubsonicMuxService) scrobbleWithScrobbler(serviceName string, id string, time_ time.Time, submission bool) error {
	if svc.scrobbler == nil {
		return nil
	}

	song, err := svc.cachedMuxSong.GetById(serviceName, id)
	if err != nil {
		return err
	}

	if submission {
		return svc.scrobbler.ScrobbleCompleted(time_, song.Artist, song.Album, song.Title)
	} else {
		return svc.scrobbler.ScrobblePlaying(song.Artist, song.Album, song.Title)
	}
}

func (svc *SubsonicMuxService) GetCoverArt(id string) (mime string, reader io.ReadCloser, err error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return
	}

	return service.GetCoverArt(id)
}

func (svc *SubsonicMuxService) Stream(ctx context.Context, id string) (mime string, reader io.ReadCloser, err error) {
	service, err := svc.findServiceByEntityId(id)
	if err != nil {
		return
	}

	return service.Stream(ctx, id)
}

func (svc *SubsonicMuxService) findServiceByName(name string) (*SubsonicNamedService, error) {
	for _, service := range svc.services {
		if service.Name() == name {
			return service, nil
		}
	}

	return nil, fmt.Errorf("unknown subsonic service `%s`", name)
}

func (svc *SubsonicMuxService) findServiceByEntityId(id string) (*SubsonicNamedService, error) {
	for _, service := range svc.services {
		if service.Matches(id) {
			return service, nil
		}
	}

	return nil, fmt.Errorf("failed to find the backing subsonic service for id `%s`", id)
}
