package tasks

import (
	"fmt"
	"log/slog"
	"tapesonic/logic"
	"tapesonic/storage"
	"time"
)

type SyncLibraryHandler struct {
	subsonicProviders []*logic.SubsonicNamedService

	songs   *storage.CachedMuxSongStorage
	albums  *storage.CachedMuxAlbumStorage
	artists *storage.CachedMuxArtistStorage
}

func NewSyncLibraryHandler(
	subsonicProviders []*logic.SubsonicNamedService,

	songs *storage.CachedMuxSongStorage,
	albums *storage.CachedMuxAlbumStorage,
	artists *storage.CachedMuxArtistStorage,
) *SyncLibraryHandler {
	return &SyncLibraryHandler{
		subsonicProviders: subsonicProviders,
		songs:             songs,
		albums:            albums,
		artists:           artists,
	}
}

func (h *SyncLibraryHandler) Name() string {
	return "SYNC_LIBRARY"
}

func (h *SyncLibraryHandler) OnSchedule() error {
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
				return fmt.Errorf("failed to make a no-query search while syncing the library cache, aborting: %w", err)
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
						Album:       song.Album,
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
		return fmt.Errorf("failed to save refreshed artist cache: %w", err)
	}

	err = h.albums.Replace(albums)
	if err != nil {
		return fmt.Errorf("failed to save refreshed album cache: %w", err)
	}

	err = h.songs.Replace(songs)
	if err != nil {
		return fmt.Errorf("failed to save refreshed song cache: %w", err)
	}

	slog.Info("Done refreshing the library cache")
	return nil
}
