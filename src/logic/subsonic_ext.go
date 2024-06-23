package logic

import (
	"context"
	"io"
	"tapesonic/http/subsonic/client"
	"tapesonic/http/subsonic/responses"
	"time"
)

type subsonicExternalService struct {
	client *client.SubsonicClient
}

func NewSubsonicExternalService(client *client.SubsonicClient) SubsonicService {
	return &subsonicExternalService{
		client: client,
	}
}

func (svc *subsonicExternalService) GetSong(id string) (*responses.SubsonicChild, error) {
	return svc.client.GetSong(id)
}

func (svc *subsonicExternalService) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	return svc.client.GetRandomSongs(size, genre, fromYear, toYear)
}

func (svc *subsonicExternalService) GetAlbum(id string) (*responses.AlbumId3, error) {
	return svc.client.GetAlbum(id)
}

func (svc *subsonicExternalService) GetAlbumList2(type_ string, size int, offset int, fromYear *int, toYear *int) (*responses.AlbumList2, error) {
	return svc.client.GetAlbumList2(type_, size, offset, fromYear, toYear)
}

func (svc *subsonicExternalService) GetPlaylist(id string) (*responses.SubsonicPlaylist, error) {
	return svc.client.GetPlaylist(id)
}

func (svc *subsonicExternalService) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	return svc.client.GetPlaylists()
}

func (svc *subsonicExternalService) Scrobble(id string, time_ time.Time, submission bool) error {
	return svc.client.Scrobble(id, time_, submission)
}

func (svc *subsonicExternalService) GetCoverArt(id string) (mime string, reader io.ReadCloser, err error) {
	return svc.client.GetCoverArt(id)
}

func (svc *subsonicExternalService) Stream(ctx context.Context, id string) (mime string, reader io.ReadCloser, err error) {
	return svc.client.Stream(id)
}
