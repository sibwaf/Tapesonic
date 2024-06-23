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

func (svc *subsonicInternalService) GetSong(rawId string) (*responses.SubsonicChild, error) {
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	track, err := svc.tracks.GetSubsonicTrack(id)
	if err != nil {
		return nil, err
	}

	songResponse := responses.NewSubsonicChild(
		track.Id.String(),
		false,
		track.Artist,
		track.Title,
		0,
		track.DurationSec,
	)
	songResponse.PlayCount = track.PlayCount

	if track.AlbumId != uuid.Nil {
		songResponse.Album = track.Album
		songResponse.AlbumId = track.AlbumId.String()
		songResponse.CoverArt = getAlbumCoverId(track.AlbumId)
	}

	return songResponse, nil
}

func (svc *subsonicInternalService) GetAlbum(rawId string) (*responses.AlbumId3, error) {
	id, err := uuid.Parse(rawId)
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

	albumResponse := responses.NewAlbumId3(
		fmt.Sprint(album.Id),
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

	for _, track := range tracks {
		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.Id),
			false,
			track.Artist,
			track.Title,
			track.AlbumTrackIndex+1,
			track.DurationSec,
		)
		trackResponse.Album = album.Name
		trackResponse.AlbumId = album.Id.String()
		trackResponse.CoverArt = getAlbumCoverId(album.Id)
		trackResponse.PlayCount = track.PlayCount

		albumResponse.Song = append(albumResponse.Song, *trackResponse)
	}

	return albumResponse, nil
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
		albumResponse := responses.NewAlbumId3(
			fmt.Sprint(album.Id),
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

		albumsResponse = append(albumsResponse, *albumResponse)
	}

	return responses.NewAlbumList2(albumsResponse), nil
}

func (svc *subsonicInternalService) GetPlaylist(rawId string) (*responses.SubsonicPlaylist, error) {
	id, err := uuid.Parse(rawId)
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

	playlistResponse := responses.NewSubsonicPlaylist(
		fmt.Sprint(playlist.Id),
		playlist.Name,
		playlist.SongCount,
		playlist.DurationSec,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	)
	playlistResponse.CoverArt = getPlaylistCoverId(playlist.Id)

	for _, track := range tracks {
		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.Id),
			false,
			track.Artist,
			track.Title,
			track.PlaylistTrackIndex+1,
			track.DurationSec,
		)
		trackResponse.CoverArt = getPlaylistCoverId(playlist.Id)
		trackResponse.PlayCount = track.PlayCount

		if track.AlbumId != uuid.Nil {
			trackResponse.Album = track.Album
			trackResponse.AlbumId = track.AlbumId.String()
			trackResponse.CoverArt = getAlbumCoverId(track.AlbumId)
		}

		playlistResponse.Entry = append(playlistResponse.Entry, *trackResponse)
	}

	return playlistResponse, nil
}

func (svc *subsonicInternalService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists, err := svc.playlists.GetSubsonicPlaylists(math.MaxInt32, 0)
	if err != nil {
		return nil, err
	}

	playlistsResponse := []responses.SubsonicPlaylist{}
	for _, playlist := range playlists {
		responsePlaylist := responses.NewSubsonicPlaylist(
			fmt.Sprint(playlist.Id),
			playlist.Name,
			playlist.SongCount,
			playlist.DurationSec,
			playlist.CreatedAt,
			playlist.UpdatedAt,
		)
		responsePlaylist.CoverArt = getPlaylistCoverId(playlist.Id)

		playlistsResponse = append(playlistsResponse, *responsePlaylist)
	}

	return responses.NewSubsonicPlaylists(playlistsResponse), nil
}

func getAlbumCoverId(albumId uuid.UUID) string {
	return "album/" + fmt.Sprint(albumId)
}

func getPlaylistCoverId(playlistId uuid.UUID) string {
	return "playlist/" + fmt.Sprint(playlistId)
}

func (svc *subsonicInternalService) Scrobble(rawId string, time_ time.Time, submission bool) error {
	id, err := uuid.Parse(rawId)
	if err != nil {
		return err
	}

	return svc.listens.Record(id, time_, submission)
}

func (svc *subsonicInternalService) GetCoverArt(rawId string) (mime string, reader io.ReadCloser, err error) {
	var cover storage.CoverDescriptor
	if strings.HasPrefix(rawId, "playlist/") {
		id, e := uuid.Parse(strings.TrimPrefix(rawId, "playlist/"))
		if e != nil {
			err = e
		} else {
			cover, err = svc.media.GetPlaylistCover(id)
		}
	} else if strings.HasPrefix(rawId, "album/") {
		id, e := uuid.Parse(strings.TrimPrefix(rawId, "album/"))
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
	id, err := uuid.Parse(rawId)
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
