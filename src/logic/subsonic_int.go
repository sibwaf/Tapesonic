package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"slices"
	"strings"
	"tapesonic/ffmpeg"
	"tapesonic/http/subsonic/responses"
	"tapesonic/storage"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type subsonicInternalService struct {
	tracks      *storage.TrackStorage
	albums      *storage.AlbumStorage
	playlists   *storage.PlaylistStorage
	listens     *storage.TrackListensStorage
	media       *storage.MediaStorage
	streamCache *storage.StreamCacheStorage

	ffmpeg *ffmpeg.Ffmpeg

	thumbnails *ThumbnailService
	scrobbler  *ScrobbleService
	ytdlp      *YtdlpService
}

func NewSubsonicInternalService(
	tracks *storage.TrackStorage,
	albums *storage.AlbumStorage,
	playlists *storage.PlaylistStorage,
	listens *storage.TrackListensStorage,
	media *storage.MediaStorage,
	streamCache *storage.StreamCacheStorage,
	ffmpeg *ffmpeg.Ffmpeg,
	thumbnails *ThumbnailService,
	scrobbler *ScrobbleService,
	ytdlp *YtdlpService,
) SubsonicService {
	return &subsonicInternalService{
		tracks:      tracks,
		albums:      albums,
		playlists:   playlists,
		listens:     listens,
		media:       media,
		streamCache: streamCache,
		ffmpeg:      ffmpeg,
		thumbnails:  thumbnails,
		scrobbler:   scrobbler,
		ytdlp:       ytdlp,
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
		if albums, err = svc.albums.SearchSubsonicAlbums(albumCount, albumOffset, query); err != nil {
			return nil, err
		}
		if songs, err = svc.tracks.SearchSubsonicTracks(songCount, songOffset, query); err != nil {
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
	coverArtId := ""
	if album.ThumbnailId != nil {
		coverArtId = encodeId(album.ThumbnailId.String())
	}

	albumResponse := responses.NewAlbumId3(
		encodeId(album.Id),
		album.Name,
		album.Artist,
		coverArtId,
		album.SongCount,
		album.DurationSec,
		album.CreatedAt,
	)

	if album.ReleaseDate != nil {
		albumResponse.Year = album.ReleaseDate.Year()
		albumResponse.ReleaseDate = responses.NewItemDate(
			album.ReleaseDate.Year(),
			int(album.ReleaseDate.Month()),
			album.ReleaseDate.Day(),
		)
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

	if playlist.ThumbnailId != nil {
		responsePlaylist.CoverArt = encodeId(playlist.ThumbnailId.String())
	}

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

	if track.AlbumId != "" {
		trackResponse.AlbumId = encodeId(track.AlbumId)
	}

	if track.AlbumThumbnailId != nil {
		trackResponse.CoverArt = encodeId(track.AlbumThumbnailId.String())
	} else if track.ThumbnailId != nil {
		trackResponse.CoverArt = encodeId(track.ThumbnailId.String())
	}

	if track.Album != "" {
		trackResponse.Album = track.Album
	}

	trackResponse.PlayCount = track.PlayCount

	return *trackResponse
}

func (svc *subsonicInternalService) Scrobble(rawId string, time_ time.Time, submission bool) error {
	id, err := decodeId(rawId)
	if err != nil {
		return err
	}

	selfErr := svc.listens.Record(id, time_, submission)
	scrobblerError := svc.scrobbleWithScrobbler(rawId, time_, submission)

	return errors.Join(selfErr, scrobblerError)
}

func (svc *subsonicInternalService) scrobbleWithScrobbler(rawId string, time_ time.Time, submission bool) error {
	if svc.scrobbler == nil {
		return nil
	}

	song, err := svc.GetSong(rawId)
	if err != nil {
		return err
	}

	if submission {
		return svc.scrobbler.ScrobbleCompleted(time_, song.Artist, song.Album, song.Title)
	} else {
		return svc.scrobbler.ScrobblePlaying(song.Artist, song.Album, song.Title)
	}
}

func (svc *subsonicInternalService) GetCoverArt(rawId string) (mediaType string, reader io.ReadCloser, err error) {
	id, err := decodeId(rawId)
	if err != nil {
		return "", nil, err
	}

	return svc.thumbnails.GetThumbnailContent(id)
}

// some codecs like mp4/alac are not supported by Chromium-based clients
var ALLOWED_STREAMING_CODECS = []string{"mp3", "flac", "opus"}

func (svc *subsonicInternalService) Stream(ctx context.Context, rawId string) (AudioStream, error) {
	id, err := decodeId(rawId)
	if err != nil {
		return AudioStream{}, err
	}

	track, err := svc.media.GetTrackSources(id)
	if err != nil {
		return AudioStream{}, err
	}

	if track.LocalPath != "" {
		allowDirectStreaming := true
		switch {
		case track.StartOffsetMs > 0:
			slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because StartOffsetMs > 0 (%d)", id, track.LocalPath, track.StartOffsetMs))
			allowDirectStreaming = false
		case track.EndOffsetMs != track.SourceDurationMs:
			slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because EndOffsetMs != SourceDurationMs (%d != %d)", id, track.LocalPath, track.EndOffsetMs, track.SourceDurationMs))
			allowDirectStreaming = false
		case !slices.Contains(ALLOWED_STREAMING_CODECS, track.LocalCodec):
			slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because codec `%s` is not allowed", id, track.LocalPath, track.LocalCodec))
			allowDirectStreaming = false
		}

		if allowDirectStreaming {
			slog.Debug(fmt.Sprintf("Streaming downloaded track id=`%s` (%s) directly from file", id, track.LocalPath))

			reader, err := os.Open(track.LocalPath)
			if err != nil {
				return AudioStream{}, err
			}

			slog.Debug(fmt.Sprintf("Got streaming data for track id=`%s`", id))

			return AudioStream{
				Reader:   reader,
				MimeType: util.FormatToMediaType(track.LocalFormat),
			}, nil
		} else {
			slog.Debug(fmt.Sprintf("Streaming downloaded track id=`%s` (%s) via ffmpeg, start=%d, end=%d", id, track.LocalPath, track.StartOffsetMs, track.EndOffsetMs))

			item, reader, err := svc.streamCache.GetOrSave(fmt.Sprintf("tapesonic-%s", id), func() (string, io.ReadCloser, error) {
				slog.Debug(fmt.Sprintf("Populating stream cache for track id=`%s`", id))

				format, reader, err := svc.ffmpeg.StreamFrom(
					ctx,
					track.LocalCodec,
					ffmpeg.ANY_FORMAT,
					track.StartOffsetMs,
					track.EndOffsetMs-track.StartOffsetMs,
					track.LocalPath,
				)
				if err != nil {
					return "", nil, err
				}

				return util.FormatToMediaType(format), reader, nil
			})
			if err != nil {
				return AudioStream{}, err
			}

			slog.Debug(fmt.Sprintf("Got streaming data for track id=`%s`", id))

			return AudioStream{
				Reader:   reader,
				MimeType: item.ContentType,
			}, nil
		}
	} else if track.RemoteUrl != "" {
		streamInfo, err := svc.ytdlp.GetStreamInfo(ctx, track.RemoteUrl, "ba")
		if err != nil {
			return AudioStream{}, err
		}

		slog.Debug(fmt.Sprintf("Streaming remote track id=`%s` (%s) via ffmpeg, start=%d, end=%d", id, track.RemoteUrl, track.StartOffsetMs, track.EndOffsetMs))

		format, reader, err := svc.ffmpeg.StreamFrom(
			ctx,
			streamInfo.ACodec,
			ffmpeg.SEEKABLE_FORMAT,
			track.StartOffsetMs,
			track.EndOffsetMs-track.StartOffsetMs,
			streamInfo.Url,
		)
		if err != nil {
			return AudioStream{}, err
		}

		slog.Debug(fmt.Sprintf("Got streaming data for track id=`%s`", id))

		return AudioStream{
			Reader:   reader,
			MimeType: util.FormatToMediaType(format),
		}, nil
	} else {
		return AudioStream{}, fmt.Errorf("no local path or remote url for track id=`%s`", id)
	}
}

func (svc *subsonicInternalService) GetLicense() (*responses.License, error) {
	return responses.NewLicense(true), nil
}

func encodeId(id string) string {
	return strings.ReplaceAll(fmt.Sprint(id), "-", "_")
}

func decodeId(rawId string) (uuid.UUID, error) {
	return uuid.Parse(strings.ReplaceAll(rawId, "_", "-"))
}
