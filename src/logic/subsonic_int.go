package logic

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	albums    *storage.AlbumStorage
	playlists *storage.PlaylistStorage
	listens   *storage.TapeTrackListensStorage
	media     *storage.MediaStorage

	ffmpeg *ffmpeg.Ffmpeg
}

func NewSubsonicInternalService(
	albums *storage.AlbumStorage,
	playlists *storage.PlaylistStorage,
	listens *storage.TapeTrackListensStorage,
	media *storage.MediaStorage,
	ffmpeg *ffmpeg.Ffmpeg,
) SubsonicService {
	return &subsonicInternalService{
		albums:    albums,
		playlists: playlists,
		listens:   listens,
		media:     media,
		ffmpeg:    ffmpeg,
	}
}

func (svc *subsonicInternalService) GetAlbum(rawId string) (*responses.AlbumId3, error) {
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	album, err := svc.albums.GetAlbumWithTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for index, track := range album.Tracks {
		lengthMs := track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs

		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.TapeTrack.Id),
			false,
			track.TapeTrack.Artist,
			track.TapeTrack.Title,
			index+1,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
	}

	albumResponse := responses.NewAlbumId3(
		fmt.Sprint(album.Id),
		album.Name,
		album.Artist,
		"album/"+fmt.Sprint(album.Id),
		len(album.Tracks),
		totalLengthMs/1000,
		album.CreatedAt,
	)
	albumResponse.Song = tracks

	return albumResponse, nil
}

func (svc *subsonicInternalService) GetAlbumList2(
	type_ string,
	size int,
	offset int,
) (*responses.AlbumList2, error) {
	var albums []storage.SubsonicAlbumListItem
	var err error
	if type_ == LIST_RANDOM {
		albums, err = svc.albums.GetSubsonicAlbumsSortRandom(size, offset)
	} else if type_ == LIST_NEWEST {
		albums, err = svc.albums.GetSubsonicAlbumsSortNewest(size, offset)
	} else if type_ == LIST_BY_NAME {
		albums, err = svc.albums.GetSubsonicAlbumsSortName(size, offset)
	} else if type_ == LIST_BY_ARTIST {
		albums, err = svc.albums.GetSubsonicAlbumsSortArtist(size, offset)
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
			"album/"+fmt.Sprint(album.Id),
			album.SongCount,
			album.DurationSec,
			album.CreatedAt,
		)
		albumsResponse = append(albumsResponse, *albumResponse)
	}

	return responses.NewAlbumList2(albumsResponse), nil
}

func (svc *subsonicInternalService) GetPlaylist(rawId string) (*responses.SubsonicPlaylist, error) {
	id, err := uuid.Parse(rawId)
	if err != nil {
		return nil, err
	}

	playlist, err := svc.playlists.GetPlaylistWithTracks(id)
	if err != nil {
		return nil, err
	}

	tracks := []responses.SubsonicChild{}
	totalLengthMs := 0
	for index, track := range playlist.Tracks {
		lengthMs := track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs

		trackResponse := responses.NewSubsonicChild(
			fmt.Sprint(track.TapeTrack.Id),
			false,
			track.TapeTrack.Artist,
			track.TapeTrack.Title,
			index+1,
			lengthMs/1000,
		)

		tracks = append(tracks, *trackResponse)
		totalLengthMs += lengthMs
	}

	playlistResponse := responses.NewSubsonicPlaylist(
		fmt.Sprint(playlist.Id),
		playlist.Name,
		len(playlist.Tracks),
		totalLengthMs/1000,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	)
	playlistResponse.CoverArt = "playlist/" + fmt.Sprint(playlist.Id)
	playlistResponse.Entry = tracks

	return playlistResponse, nil
}

func (svc *subsonicInternalService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	playlists, err := svc.playlists.GetAllPlaylists()
	if err != nil {
		return nil, err
	}

	playlistsResponse := []responses.SubsonicPlaylist{}
	for _, playlist := range playlists {
		totalLengthMs := 0
		for _, track := range playlist.Tracks {
			totalLengthMs += track.TapeTrack.EndOffsetMs - track.TapeTrack.StartOffsetMs
		}

		responsePlaylist := responses.NewSubsonicPlaylist(
			fmt.Sprint(playlist.Id),
			playlist.Name,
			len(playlist.Tracks),
			totalLengthMs/1000,
			playlist.CreatedAt,
			playlist.UpdatedAt,
		)
		responsePlaylist.CoverArt = "playlist/" + fmt.Sprint(responsePlaylist.Id)

		playlistsResponse = append(playlistsResponse, *responsePlaylist)
	}

	return responses.NewSubsonicPlaylists(playlistsResponse), nil
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
