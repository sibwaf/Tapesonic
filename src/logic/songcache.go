package logic

import (
	"fmt"
	"slices"
	"tapesonic/storage"
	"time"
)

type SongCacheService struct {
	subsonic     map[string]*SubsonicNamedService
	cache        *storage.CachedMuxSongStorage
	trackMatcher *TrackMatcher
}

func NewSongCacheService(
	subsonicServices []*SubsonicNamedService,
	cache *storage.CachedMuxSongStorage,
	trackMatcher *TrackMatcher,
) *SongCacheService {
	subsonic := make(map[string]*SubsonicNamedService)
	for _, svc := range subsonicServices {
		subsonic[svc.Name()] = svc
	}

	return &SongCacheService{
		subsonic:     subsonic,
		cache:        cache,
		trackMatcher: trackMatcher,
	}
}

func (s *SongCacheService) FindCachedSongByFields(artist string, title string, album string) (*storage.CachedMuxSong, error) {
	tracks, err := s.cache.SearchByFields(artist, album, title, 2)
	if err != nil {
		return nil, err
	}

	expected := TrackForMatching{
		Artist: artist,
		Title:  title,
	}
	tracks = slices.DeleteFunc(tracks, func(t storage.CachedMuxSong) bool {
		actual := TrackForMatching{
			Artist: t.Artist,
			Title:  t.Title,
		}
		return !s.trackMatcher.Match(expected, actual)
	})

	if len(tracks) == 1 {
		return &tracks[0], nil
	}

	// todo
	// - if nothing found, we can try omitting album as a track can be hanging around without it

	return nil, nil
}

func (s *SongCacheService) Refresh(serviceName string, id string) (storage.CachedMuxSong, error) {
	subsonic, ok := s.subsonic[serviceName]
	if !ok {
		return storage.CachedMuxSong{}, fmt.Errorf("unknown service: %s", serviceName)
	}

	song, err := subsonic.GetSongByRawId(id)
	if err != nil {
		return storage.CachedMuxSong{}, err
	}

	rawSong := subsonic.GetRawSong(*song)
	cachedSong := storage.CachedMuxSong{
		ServiceName: subsonic.Name(),
		SongId:      rawSong.Id,

		AlbumId: rawSong.AlbumId,

		Artist: rawSong.Artist,
		Album:  rawSong.Album,
		Title:  rawSong.Title,

		DurationSec: rawSong.Duration,

		CachedAt: time.Now(),
	}

	return s.cache.Save(cachedSong)
}
