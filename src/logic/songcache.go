package logic

import (
	"fmt"
	"tapesonic/storage"
	"time"
)

type SongCacheService struct {
	subsonic map[string]*SubsonicNamedService
	cache    *storage.CachedMuxSongStorage
}

func NewSongCacheService(
	subsonicServices []*SubsonicNamedService,
	cache *storage.CachedMuxSongStorage,
) *SongCacheService {
	subsonic := make(map[string]*SubsonicNamedService)
	for _, svc := range subsonicServices {
		subsonic[svc.Name()] = svc
	}

	return &SongCacheService{
		subsonic: subsonic,
		cache:    cache,
	}
}

func (s *SongCacheService) FindSongIdByFields(artist string, title string, album string) (*storage.CachedSongId, error) {
	tracks, err := s.cache.SearchByFields(artist, album, title, 2)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 1 {
		return &tracks[0], nil
	}

	// todo
	// - if nothing found, we can try omitting album as a track can be hanging around without it
	// - if multiple tracks were found, we can probably return the one with the shortest name
	// - needs a check that fields do not contain any extra words to fix matching for things like "Name" "Name Pt.2"

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
