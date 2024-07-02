package logic

import (
	"context"
	"fmt"
	"io"
	"strings"
	"tapesonic/http/subsonic/responses"
	"time"
)

type subsonicNamedService struct {
	name     string
	delegate SubsonicService
}

func NewSubsonicNamedService(name string, delegate SubsonicService) *subsonicNamedService {
	return &subsonicNamedService{
		name:     name,
		delegate: delegate,
	}
}

func (svc *subsonicNamedService) Search3(query string, artistCount int, artistOffset int, albumCount int, albumOffset int, songCount int, songOffset int) (*responses.SearchResult3, error) {
	search, err := svc.delegate.Search3(query, artistCount, artistOffset, albumCount, albumOffset, songCount, songOffset)
	if err != nil {
		return nil, err
	}

	for i := range search.Artist {
		search.Artist[i] = svc.rewriteArtistInfo(search.Artist[i])
	}
	for i := range search.Album {
		search.Album[i] = svc.rewriteAlbumInfo(search.Album[i])
	}
	for i := range search.Song {
		search.Song[i] = svc.rewriteSongInfo(search.Song[i])
	}

	return search, nil
}

func (svc *subsonicNamedService) GetSong(id string) (*responses.SubsonicChild, error) {
	song, err := svc.delegate.GetSong(svc.RemovePrefix(id))
	if err != nil {
		return nil, err
	}

	rewrittenSong := svc.rewriteSongInfo(*song)
	return &rewrittenSong, nil
}

func (svc *subsonicNamedService) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	songs, err := svc.delegate.GetRandomSongs(size, genre, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	for i := range songs.Song {
		songs.Song[i] = svc.rewriteSongInfo(songs.Song[i])
	}

	return songs, nil
}

func (svc *subsonicNamedService) GetAlbum(id string) (*responses.AlbumId3, error) {
	album, err := svc.delegate.GetAlbum(svc.RemovePrefix(id))
	if err != nil {
		return nil, err
	}

	rewrittenAlbum := svc.rewriteAlbumInfo(*album)
	return &rewrittenAlbum, nil
}

func (svc *subsonicNamedService) GetAlbumList2(type_ string, size int, offset int, fromYear *int, toYear *int) (*responses.AlbumList2, error) {
	albumList, err := svc.delegate.GetAlbumList2(type_, size, offset, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	for i := range albumList.Album {
		albumList.Album[i] = svc.rewriteAlbumInfo(albumList.Album[i])
	}

	return albumList, nil
}

func (svc *subsonicNamedService) GetPlaylist(id string) (*responses.SubsonicPlaylist, error) {
	playlist, err := svc.delegate.GetPlaylist(svc.RemovePrefix(id))
	if err != nil {
		return nil, err
	}

	rewrittenPlaylist := svc.rewritePlaylistInfo(*playlist)
	return &rewrittenPlaylist, nil
}

func (svc *subsonicNamedService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists, err := svc.delegate.GetPlaylists()
	if err != nil {
		return nil, err
	}

	for i := range playlists.Playlist {
		playlists.Playlist[i] = svc.rewritePlaylistInfo(playlists.Playlist[i])
	}

	return playlists, nil
}

func (svc *subsonicNamedService) Scrobble(id string, time_ time.Time, submission bool) error {
	return svc.delegate.Scrobble(svc.RemovePrefix(id), time_, submission)
}

func (svc *subsonicNamedService) GetCoverArt(id string) (mime string, reader io.ReadCloser, err error) {
	return svc.delegate.GetCoverArt(svc.RemovePrefix(id))
}

func (svc *subsonicNamedService) Stream(ctx context.Context, id string) (mime string, reader io.ReadCloser, err error) {
	return svc.delegate.Stream(ctx, svc.RemovePrefix(id))
}

func (svc *subsonicNamedService) rewriteAlbumInfo(album responses.AlbumId3) responses.AlbumId3 {
	album.Id = svc.addPrefix(album.Id)
	album.CoverArt = svc.addPrefix(album.CoverArt)

	for i := range album.Song {
		album.Song[i] = svc.rewriteSongInfo(album.Song[i])
	}

	return album
}

func (svc *subsonicNamedService) GetRawAlbum(album responses.AlbumId3) responses.AlbumId3 {
	album.Id = svc.RemovePrefix(album.Id)
	album.CoverArt = svc.RemovePrefix(album.CoverArt)

	for i := range album.Song {
		album.Song[i] = svc.GetRawSong(album.Song[i])
	}

	return album
}

func (svc *subsonicNamedService) rewritePlaylistInfo(playlist responses.SubsonicPlaylist) responses.SubsonicPlaylist {
	playlist.Id = svc.addPrefix(playlist.Id)
	playlist.CoverArt = svc.addPrefix(playlist.CoverArt)

	for i := range playlist.Entry {
		playlist.Entry[i] = svc.rewriteSongInfo(playlist.Entry[i])
	}

	return playlist
}

func (svc *subsonicNamedService) GetRawPlaylist(playlist responses.SubsonicPlaylist) responses.SubsonicPlaylist {
	playlist.Id = svc.RemovePrefix(playlist.Id)
	playlist.CoverArt = svc.RemovePrefix(playlist.CoverArt)

	for i := range playlist.Entry {
		playlist.Entry[i] = svc.GetRawSong(playlist.Entry[i])
	}

	return playlist
}

func (svc *subsonicNamedService) rewriteSongInfo(song responses.SubsonicChild) responses.SubsonicChild {
	song.Id = svc.addPrefix(song.Id)
	song.CoverArt = svc.addPrefix(song.CoverArt)
	song.AlbumId = svc.addPrefix(song.AlbumId)
	return song
}

func (svc *subsonicNamedService) GetRawSong(song responses.SubsonicChild) responses.SubsonicChild {
	song.Id = svc.RemovePrefix(song.Id)
	song.CoverArt = svc.RemovePrefix(song.CoverArt)
	song.AlbumId = svc.RemovePrefix(song.AlbumId)
	return song
}

func (svc *subsonicNamedService) rewriteArtistInfo(artist responses.ArtistId3) responses.ArtistId3 {
	artist.Id = svc.addPrefix(artist.Id)
	artist.CoverArt = svc.addPrefix(artist.CoverArt)
	return artist
}

func (svc *subsonicNamedService) GetRawArtist(artist responses.ArtistId3) responses.ArtistId3 {
	artist.Id = svc.RemovePrefix(artist.Id)
	artist.CoverArt = svc.RemovePrefix(artist.CoverArt)
	return artist
}

func (svc *subsonicNamedService) Name() string {
	return svc.name
}

func (svc *subsonicNamedService) Matches(id string) bool {
	return strings.HasPrefix(id, svc.generatePrefix())
}

func (svc *subsonicNamedService) addPrefix(id string) string {
	if id == "" {
		return ""
	}

	return fmt.Sprintf("%s%s", svc.generatePrefix(), id)
}

func (svc *subsonicNamedService) RemovePrefix(id string) string {
	return strings.TrimPrefix(id, svc.generatePrefix())
}

func (svc *subsonicNamedService) generatePrefix() string {
	return fmt.Sprintf("%s_", svc.name)
}
