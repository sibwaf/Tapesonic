package tasks

import (
	"fmt"
	"log/slog"
	"tapesonic/config"
	"tapesonic/logic"
	"tapesonic/storage"
	"time"

	"github.com/robfig/cron/v3"
)

type SyncLibraryHandler struct {
	subsonicProviders []*logic.SubsonicNamedService

	songs   *storage.CachedMuxSongStorage
	albums  *storage.CachedMuxAlbumStorage
	artists *storage.CachedMuxArtistStorage

	taskConfig config.BackgroundTaskConfig
}

func NewSyncLibraryHandler(
	subsonicProviders []*logic.SubsonicNamedService,

	songs *storage.CachedMuxSongStorage,
	albums *storage.CachedMuxAlbumStorage,
	artists *storage.CachedMuxArtistStorage,

	taskConfig config.BackgroundTaskConfig,
) *SyncLibraryHandler {
	return &SyncLibraryHandler{
		subsonicProviders: subsonicProviders,
		songs:             songs,
		albums:            albums,
		artists:           artists,

		taskConfig: taskConfig,
	}
}

func (h *SyncLibraryHandler) RegisterSchedules(cron *cron.Cron) error {
	_, err := cron.AddFunc(h.taskConfig.Cron, h.onSchedule)
	return err
}

func (h *SyncLibraryHandler) onSchedule() {
	slog.Debug("Refreshing the library cache")

	artists := []storage.CachedMuxArtist{}
	albums := []storage.CachedMuxAlbum{}
	songs := []storage.CachedMuxSong{}

	batchSize := 500
	for _, subsonicProvider := range h.subsonicProviders {
		thisArtists := []storage.CachedMuxArtist{}
		thisAlbums := []storage.CachedMuxAlbum{}
		thisSongs := []storage.CachedMuxSong{}

		for {
			slog.Debug(fmt.Sprintf("Requesting %d more contents from subsonic `%s` for the library cache sync", batchSize, subsonicProvider.Name()))
			search, err := subsonicProvider.Search3("", batchSize, len(thisArtists), batchSize, len(thisAlbums), batchSize, len(thisSongs))
			if err != nil {
				slog.Error(fmt.Sprintf("Failed to make a no-query search while syncing the library cache, aborting: %s", err.Error()))
				return
			}

			slog.Debug(fmt.Sprintf("Got another %d artists, %d albums and %d songs from subsonic `%s` while syncing the library cache", len(search.Artist), len(search.Album), len(search.Song), subsonicProvider.Name()))

			cachedAt := time.Now()

			for _, artist := range search.Artist {
				artist = subsonicProvider.GetRawArtistId3(artist)
				thisArtists = append(
					thisArtists,
					storage.CachedMuxArtist{
						ServiceName: subsonicProvider.Name(),
						ArtistId:    artist.Id,
						Name:        artist.Name,
						CachedAt:    cachedAt,
					},
				)
			}

			for _, album := range search.Album {
				album = subsonicProvider.GetRawAlbum(album)
				thisAlbums = append(
					thisAlbums,
					storage.CachedMuxAlbum{
						ServiceName: subsonicProvider.Name(),
						AlbumId:     album.Id,
						Artist:      album.Artist,
						Title:       album.Name,
						CachedAt:    cachedAt,
					},
				)
			}

			for _, song := range search.Song {
				song = subsonicProvider.GetRawSong(song)
				thisSongs = append(
					thisSongs,
					storage.CachedMuxSong{
						ServiceName: subsonicProvider.Name(),
						SongId:      song.Id,
						AlbumId:     song.AlbumId,
						Artist:      song.Artist,
						Title:       song.Title,
						DurationSec: song.Duration,
						CachedAt:    cachedAt,
					},
				)
			}

			if len(search.Artist) < batchSize && len(search.Album) < batchSize && len(search.Song) < batchSize {
				slog.Debug(fmt.Sprintf("Got a total of %d artists, %d albums, %d songs from subsonic `%s`", len(thisArtists), len(thisAlbums), len(thisSongs), subsonicProvider.Name()))
				break
			}
		}

		artists = append(artists, thisArtists...)
		albums = append(albums, thisAlbums...)
		songs = append(songs, thisSongs...)
	}

	slog.Debug(fmt.Sprintf("Got a total of %d artists, %d albums, %d songs while refreshing the library cache", len(artists), len(albums), len(songs)))

	err := h.artists.Replace(artists)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to save refreshed artist cache: %s", err.Error()))
	}

	err = h.albums.Replace(albums)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to save refreshed album cache: %s", err.Error()))
	}

	err = h.songs.Replace(songs)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to save refreshed song cache: %s", err.Error()))
	}

	slog.Info("Done refreshing the library cache")
}
