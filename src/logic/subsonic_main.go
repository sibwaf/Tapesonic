package logic

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
	"tapesonic/util"
	"time"
)

type subsonicMainService struct {
	delegate          SubsonicService
	subsonicProviders []*SubsonicNamedService

	songCache   *storage.CachedMuxSongStorage
	albumCache  *storage.CachedMuxAlbumStorage
	artistCache *storage.CachedMuxArtistStorage

	externalPlaylists *storage.ExternalPlaylistStorage
}

func NewSubsonicMainService(
	delegate SubsonicService,
	subsonicProviders []*SubsonicNamedService,
	songCache *storage.CachedMuxSongStorage,
	albumCache *storage.CachedMuxAlbumStorage,
	artistCache *storage.CachedMuxArtistStorage,
	externalPlaylists *storage.ExternalPlaylistStorage,
) SubsonicService {
	return &subsonicMainService{
		delegate:          delegate,
		subsonicProviders: subsonicProviders,
		songCache:         songCache,
		albumCache:        albumCache,
		artistCache:       artistCache,
		externalPlaylists: externalPlaylists,
	}
}

func (svc *subsonicMainService) Search3(query string, artistCount int, artistOffset int, albumCount int, albumOffset int, songCount int, songOffset int) (*responses.SearchResult3, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return svc.delegate.Search3(query, artistCount, artistOffset, albumCount, albumOffset, songCount, songOffset)
	}

	artistIds, err := svc.artistCache.Search(query, artistCount, artistOffset)
	if err != nil {
		return nil, err
	}
	artists, err := util.ParallelMap(artistIds, func(item storage.CachedArtistId) (responses.ArtistId3, error) {
		subsonicProvider, err := svc.findServiceByName(item.ServiceName)
		if err != nil {
			return responses.ArtistId3{}, err
		}

		artist, err := subsonicProvider.GetArtist(item.Id)
		if err != nil {
			return responses.ArtistId3{}, err
		}

		artistId3 := responses.NewArtistId3(artist.Id, artist.Name)
		artistId3.AlbumCount = len(artist.Album)
		artistId3.ArtistImageUrl = artist.ArtistImageUrl
		artistId3.Starred = artist.Starred

		return *artistId3, nil
	})
	if err != nil {
		return nil, err
	}

	albumIds, err := svc.albumCache.Search(query, albumCount, albumOffset)
	if err != nil {
		return nil, err
	}
	albums, err := util.ParallelMap(albumIds, func(item storage.CachedAlbumId) (responses.AlbumId3, error) {
		subsonicProvider, err := svc.findServiceByName(item.ServiceName)
		if err != nil {
			return responses.AlbumId3{}, err
		}

		album, err := subsonicProvider.GetAlbumByRawId(item.Id)
		if err != nil {
			return responses.AlbumId3{}, err
		}

		return *album, nil
	})
	if err != nil {
		return nil, err
	}

	songIds, err := svc.songCache.Search(query, songCount, songOffset)
	if err != nil {
		return nil, err
	}
	songs, err := util.ParallelMap(songIds, func(item storage.CachedSongId) (responses.SubsonicChild, error) {
		subsonicProvider, err := svc.findServiceByName(item.ServiceName)
		if err != nil {
			return responses.SubsonicChild{}, err
		}

		song, err := subsonicProvider.GetSongByRawId(item.Id)
		if err != nil {
			return responses.SubsonicChild{}, err
		}

		return *song, nil
	})
	if err != nil {
		return nil, err
	}

	return responses.NewSearchResult3(artists, albums, songs), nil
}

func (svc *subsonicMainService) GetSong(id string) (*responses.SubsonicChild, error) {
	return svc.delegate.GetSong(id)
}

func (svc *subsonicMainService) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	return svc.delegate.GetRandomSongs(size, genre, fromYear, toYear)
}

func (svc *subsonicMainService) GetAlbum(id string) (*responses.AlbumId3, error) {
	return svc.delegate.GetAlbum(id)
}

func (svc *subsonicMainService) GetAlbumList2(type_ string, size int, offset int, fromYear *int, toYear *int) (*responses.AlbumList2, error) {
	return svc.delegate.GetAlbumList2(type_, size, offset, fromYear, toYear)
}

func (svc *subsonicMainService) GetPlaylist(id string) (*responses.SubsonicPlaylist, error) {
	if strings.HasPrefix(id, "external_") {
		id = strings.TrimPrefix(id, "external_")

		playlist, err := svc.externalPlaylists.GetSubsonicPlaylist(id)
		if err != nil {
			return nil, err
		}
		trackIds, err := svc.externalPlaylists.GetTracksByPlaylist(id)
		if err != nil {
			return nil, err
		}

		responsePlaylist := addPlaylistIdPrefix("external", playlist)

		responsePlaylist.Entry, err = util.ParallelMap(trackIds, func(item storage.CachedSongIdWithIndex) (responses.SubsonicChild, error) {
			subsonicProvider, err := svc.findServiceByName(item.ServiceName)
			if err != nil {
				return responses.SubsonicChild{}, err
			}

			song, err := subsonicProvider.GetSongByRawId(item.Id)
			if err != nil {
				return responses.SubsonicChild{}, err
			}

			song.Track = item.TrackIndex

			return *song, nil
		})
		if err != nil {
			return nil, err
		}

		return responsePlaylist, nil
	}

	return svc.delegate.GetPlaylist(id)
}

func (svc *subsonicMainService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	allPlaylists := []responses.SubsonicPlaylist{}

	externalPlaylists, err := svc.getExternalPlaylists()
	if err != nil {
		return nil, err
	}
	allPlaylists = append(allPlaylists, externalPlaylists.Playlist...)

	libraryPlaylists, err := svc.delegate.GetPlaylists()
	if err != nil {
		return nil, err
	}
	allPlaylists = append(allPlaylists, libraryPlaylists.Playlist...)

	return responses.NewSubsonicPlaylists(allPlaylists), nil
}

func (svc *subsonicMainService) getExternalPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists, err := svc.externalPlaylists.GetSubsonicPlaylists(math.MaxInt32, 0)
	if err != nil {
		return nil, err
	}

	resultPlaylists := []responses.SubsonicPlaylist{}
	for _, playlist := range playlists {
		if playlist.SongCount <= 0 {
			continue
		}

		resultPlaylists = append(resultPlaylists, *addPlaylistIdPrefix("external", playlist))
	}

	return responses.NewSubsonicPlaylists(resultPlaylists), nil
}

func addPlaylistIdPrefix(prefix string, playlist storage.SubsonicPlaylistItem) *responses.SubsonicPlaylist {
	result := responses.NewSubsonicPlaylist(
		fmt.Sprintf("%s_%s", prefix, playlist.Id),
		playlist.Name,
		playlist.SongCount,
		playlist.DurationSec,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	)

	result.Owner = playlist.CreatedBy

	return result
}

func (svc *subsonicMainService) GetArtist(id string) (*responses.Artist, error) {
	return svc.delegate.GetArtist(id)
}

func (svc *subsonicMainService) Scrobble(id string, time_ time.Time, submission bool) error {
	return svc.delegate.Scrobble(id, time_, submission)
}

func (svc *subsonicMainService) GetCoverArt(id string) (mime string, reader io.ReadCloser, err error) {
	return svc.delegate.GetCoverArt(id)
}

func (svc *subsonicMainService) Stream(ctx context.Context, id string) (AudioStream, error) {
	return svc.delegate.Stream(ctx, id)
}

func (svc *subsonicMainService) GetLicense() (*responses.License, error) {
	return svc.delegate.GetLicense()
}

func (svc *subsonicMainService) findServiceByName(name string) (*SubsonicNamedService, error) {
	for _, service := range svc.subsonicProviders {
		if service.Name() == name {
			return service, nil
		}
	}

	return nil, fmt.Errorf("unknown subsonic service `%s`", name)
}
