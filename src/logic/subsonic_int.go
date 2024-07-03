package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"tapesonic/ffmpeg"
	"tapesonic/http/subsonic/responses"
	httpUtil "tapesonic/http/util"
	"tapesonic/storage"
	commonUtil "tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type subsonicInternalService struct {
	tracks    *storage.TrackStorage
	albums    *storage.AlbumStorage
	playlists *storage.PlaylistStorage
	listens   *storage.TapeTrackListensStorage
	media     *storage.MediaStorage

	ffmpeg *ffmpeg.Ffmpeg
}

func NewSubsonicInternalService(
	tracks *storage.TrackStorage,
	albums *storage.AlbumStorage,
	playlists *storage.PlaylistStorage,
	listens *storage.TapeTrackListensStorage,
	media *storage.MediaStorage,
	ffmpeg *ffmpeg.Ffmpeg,
) SubsonicService {
	return &subsonicInternalService{
		tracks:    tracks,
		albums:    albums,
		playlists: playlists,
		listens:   listens,
		media:     media,
		ffmpeg:    ffmpeg,
	}
}

func (svc *subsonicInternalService) Search3(
	query string,
	artistCount int,
	artistOffset int,
	albumCount int,
	albumOffset int,
	songCount int,
	songOffset int,
) (*responses.SearchResult3, error) {
	var err error = nil

	var albums []storage.SubsonicAlbumItem
	var songs []storage.SubsonicTrackItem

	query = strings.TrimSpace(query)
	if query == "" {
		if albums, err = svc.albums.GetSubsonicAlbumsSortId(albumCount, albumOffset); err != nil {
			return nil, err
		}
		if songs, err = svc.tracks.GetSubsonicTracksSortId(songCount, songOffset); err != nil {
			return nil, err
		}
	} else {
		terms := []string{}
		for _, term := range strings.Split(query, " ") {
			if term != "" {
				terms = append(terms, term)
			}
		}

		if albums, err = svc.albums.SearchSubsonicAlbums(albumCount, albumOffset, terms); err != nil {
			return nil, err
		}
		if songs, err = svc.tracks.SearchSubsonicTracks(songCount, songOffset, terms); err != nil {
			return nil, err
		}
	}

	albumsResponse := []responses.AlbumId3{}
	for _, album := range albums {
		albumsResponse = append(albumsResponse, toAlbumId3(album))
	}

	songsResponse := []responses.SubsonicChild{}
	for _, song := range songs {
		songsResponse = append(songsResponse, toChild(song))
	}

	return &responses.SearchResult3{
		Artist: []responses.ArtistId3{}, // todo
		Album:  albumsResponse,
		Song:   songsResponse,
	}, nil
}

func (svc *subsonicInternalService) GetSong(rawId string) (*responses.SubsonicChild, error) {
	id, err := decodeId(rawId)
	if err != nil {
		return nil, err
	}

	track, err := svc.tracks.GetSubsonicTrack(id)
	if err != nil {
		return nil, err
	}

	songResponse := toChild(*track)
	songResponse.Track = 0

	return &songResponse, nil
}

func (svc *subsonicInternalService) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	if genre != "" {
		// todo
		return responses.NewRandomSongs([]responses.SubsonicChild{}), nil
	}

	songs, err := svc.tracks.GetSubsonicTracksSortRandom(size, fromYear, toYear)
	if err != nil {
		return nil, err
	}

	songsResponse := []responses.SubsonicChild{}
	for _, song := range songs {
		songResponse := toChild(song)
		songResponse.Track = 0

		songsResponse = append(songsResponse, songResponse)
	}

	return responses.NewRandomSongs(songsResponse), nil
}

func (svc *subsonicInternalService) GetAlbum(rawId string) (*responses.AlbumId3, error) {
	id, err := decodeId(rawId)
	if err != nil {
		return nil, err
	}

	album, err := svc.albums.GetSubsonicAlbum(id)
	if err != nil {
		return nil, err
	}

	tracks, err := svc.tracks.GetSubsonicTracksByAlbum(id)
	if err != nil {
		return nil, err
	}

	albumResponse := toAlbumId3(*album)

	for _, track := range tracks {
		albumResponse.Song = append(albumResponse.Song, toChild(track))
	}

	return &albumResponse, nil
}

func (svc *subsonicInternalService) GetAlbumList2(
	type_ string,
	size int,
	offset int,
	fromYear *int,
	toYear *int,
) (*responses.AlbumList2, error) {
	var albums []storage.SubsonicAlbumItem
	var err error
	if type_ == LIST_RANDOM {
		albums, err = svc.albums.GetSubsonicAlbumsSortRandom(size, offset)
	} else if type_ == LIST_NEWEST {
		albums, err = svc.albums.GetSubsonicAlbumsSortNewest(size, offset)
	} else if type_ == LIST_BY_NAME {
		albums, err = svc.albums.GetSubsonicAlbumsSortName(size, offset)
	} else if type_ == LIST_BY_ARTIST {
		albums, err = svc.albums.GetSubsonicAlbumsSortArtist(size, offset)
	} else if type_ == LIST_RECENT {
		albums, err = svc.albums.GetSubsonicAlbumsSortRecent(size, offset)
	} else if type_ == LIST_FREQUENT {
		albums, err = svc.albums.GetSubsonicAlbumsSortFrequent(size, offset)
	} else if type_ == LIST_STARRED {
		albums, err = []storage.SubsonicAlbumItem{}, nil
	} else if type_ == LIST_BY_YEAR {
		if fromYear == nil || toYear == nil {
			return nil, fmt.Errorf("fromYear or toYear parameter missing")
		}

		albums, err = svc.albums.GetSubsonicAlbumsSortReleaseDate(size, offset, *fromYear, *toYear)
	} else {
		return nil, fmt.Errorf("unsupported album sort order %s", type_)
	}

	if err != nil {
		return nil, err
	}

	albumsResponse := []responses.AlbumId3{}
	for _, album := range albums {
		albumsResponse = append(albumsResponse, toAlbumId3(album))
	}

	return responses.NewAlbumList2(albumsResponse), nil
}

func (svc *subsonicInternalService) GetPlaylist(rawId string) (*responses.SubsonicPlaylist, error) {
	id, err := decodeId(rawId)
	if err != nil {
		return nil, err
	}

	playlist, err := svc.playlists.GetSubsonicPlaylist(id)
	if err != nil {
		return nil, err
	}

	tracks, err := svc.tracks.GetSubsonicTracksByPlaylist(id)
	if err != nil {
		return nil, err
	}

	playlistResponse := toPlaylist(*playlist)

	for _, track := range tracks {
		trackResponse := toChild(track)

		trackResponse.Track = track.PlaylistTrackIndex + 1
		if trackResponse.CoverArt == "" {
			trackResponse.CoverArt = getPlaylistCoverId(playlist.Id)
		}

		playlistResponse.Entry = append(playlistResponse.Entry, trackResponse)
	}

	return &playlistResponse, nil
}

func (svc *subsonicInternalService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists, err := svc.playlists.GetSubsonicPlaylists(math.MaxInt32, 0)
	if err != nil {
		return nil, err
	}

	playlistsResponse := []responses.SubsonicPlaylist{}
	for _, playlist := range playlists {
		playlistsResponse = append(playlistsResponse, toPlaylist(playlist))
	}

	return responses.NewSubsonicPlaylists(playlistsResponse), nil
}

func (svc *subsonicInternalService) GetArtist(id string) (*responses.Artist, error) {
	// todo
	return nil, fmt.Errorf("not supported yet")
}

func toAlbumId3(album storage.SubsonicAlbumItem) responses.AlbumId3 {
	albumResponse := responses.NewAlbumId3(
		encodeId(album.Id),
		album.Name,
		album.Artist,
		getAlbumCoverId(album.Id),
		album.SongCount,
		album.DurationSec,
		album.CreatedAt,
	)

	if album.ReleaseDate != nil {
		albumResponse.Year = album.ReleaseDate.Year()
	}

	albumResponse.PlayCount = album.PlayCount

	return *albumResponse
}

func toPlaylist(playlist storage.SubsonicPlaylistItem) responses.SubsonicPlaylist {
	responsePlaylist := responses.NewSubsonicPlaylist(
		encodeId(playlist.Id),
		playlist.Name,
		playlist.SongCount,
		playlist.DurationSec,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	)

	responsePlaylist.CoverArt = getPlaylistCoverId(playlist.Id)

	return *responsePlaylist
}

func toChild(track storage.SubsonicTrackItem) responses.SubsonicChild {
	trackResponse := responses.NewSubsonicChild(
		encodeId(track.Id),
		false,
		track.Artist,
		track.Title,
		track.AlbumTrackIndex+1,
		track.DurationSec,
	)

	if track.AlbumId != uuid.Nil {
		trackResponse.Album = track.Album
		trackResponse.AlbumId = encodeId(track.AlbumId)
		trackResponse.CoverArt = getAlbumCoverId(track.AlbumId)
	}

	trackResponse.PlayCount = track.PlayCount

	return *trackResponse
}

func getAlbumCoverId(albumId uuid.UUID) string {
	if albumId == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("album_%s", encodeId(albumId))
}

func getPlaylistCoverId(playlistId uuid.UUID) string {
	if playlistId == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("playlist_%s", encodeId(playlistId))
}

func (svc *subsonicInternalService) Scrobble(rawId string, time_ time.Time, submission bool) error {
	id, err := decodeId(rawId)
	if err != nil {
		return err
	}

	return svc.listens.Record(id, time_, submission)
}

func (svc *subsonicInternalService) GetCoverArt(rawId string) (mime string, reader io.ReadCloser, err error) {
	var cover storage.CoverDescriptor
	if strings.HasPrefix(rawId, "playlist_") {
		id, e := decodeId(strings.TrimPrefix(rawId, "playlist_"))
		if e != nil {
			err = e
		} else {
			cover, err = svc.media.GetPlaylistCover(id)
		}
	} else if strings.HasPrefix(rawId, "album_") {
		id, e := decodeId(strings.TrimPrefix(rawId, "album_"))
		if e != nil {
			err = e
		} else {
			cover, err = svc.media.GetAlbumCover(id)
		}
	} else {
		err = fmt.Errorf("failed to find cover art `%s`", rawId)
	}

	if err != nil {
		return
	}

	mime = httpUtil.FormatToMediaType(cover.Format)
	reader, err = os.Open(cover.Path)
	return
}

func (svc *subsonicInternalService) Stream(ctx context.Context, rawId string) (mime string, reader io.ReadCloser, err error) {
	id, err := decodeId(rawId)
	if err != nil {
		return
	}

	track, err := svc.media.GetTrack(id)
	if err != nil {
		return
	}

	mime = httpUtil.FormatToMediaType(track.Format)

	sourceReader, err := os.Open(track.Path)
	if err != nil {
		return
	}

	streamReader, err := svc.ffmpeg.Stream(
		ctx,
		track.StartOffsetMs,
		track.EndOffsetMs-track.StartOffsetMs,
		ffmpeg.NewReaderWithMeta(
			"file://"+track.Path,
			sourceReader,
		),
	)
	if err != nil {
		sourceReader.Close()
		return
	}

	return mime, commonUtil.NewCustomReadCloser(streamReader, func() error {
		return errors.Join(
			streamReader.Close(),
			sourceReader.Close(),
		)
	}), nil
}

func encodeId(id uuid.UUID) string {
	return strings.ReplaceAll(fmt.Sprint(id), "-", "_")
}

func decodeId(rawId string) (uuid.UUID, error) {
	return uuid.Parse(strings.ReplaceAll(rawId, "_", "-"))
}
