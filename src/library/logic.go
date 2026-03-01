package library

import (
	"slices"
	"tapesonic/model"
	"tapesonic/users"

	"github.com/google/uuid"
)

type LibraryService struct {
	artists   *ArtistStorage
	albums    *AlbumStorage
	tracks    *TrackStorage
	covers    *CoverStorage
	playlists *PlaylistStorage
}

func newLibraryService(
	artists *ArtistStorage,
	albums *AlbumStorage,
	tracks *TrackStorage,
	covers *CoverStorage,
	playlists *PlaylistStorage,
) *LibraryService {
	return &LibraryService{
		artists:   artists,
		albums:    albums,
		tracks:    tracks,
		covers:    covers,
		playlists: playlists,
	}
}

func (svc *LibraryService) GetCover(id string) (model.LibraryCover, error) {
	cover, err := svc.covers.FindCoverById(id)
	if err != nil {
		return model.LibraryCover{}, err
	}
	if cover == nil {
		return model.LibraryCover{}, model.ErrNotFound
	}

	return *cover, nil
}

func (svc *LibraryService) GetArtist(user users.User, id string) (model.LibraryArtist, error) {
	artist, err := svc.artists.FindArtistById(user.Id, id)
	if err != nil {
		return model.LibraryArtist{}, err
	}
	if artist == nil {
		return model.LibraryArtist{}, model.ErrNotFound
	}

	albums, err := svc.albums.GetAlbumsByArtistId(user.Id, artist.Id)
	if err != nil {
		return model.LibraryArtist{}, err
	}

	artist.Albums = albums

	return *artist, err
}

func (svc *LibraryService) SearchArtistsByQuery(user users.User, query string, count int, offset int) ([]model.LibraryArtist, error) {
	return svc.artists.SearchArtistsByQuery(user.Id, query, count, offset)
}

func (svc *LibraryService) GetArtistsSortById(user users.User, count int, offset int) ([]model.LibraryArtist, error) {
	return svc.artists.GetArtistsSortId(user.Id, count, offset)
}

func (svc *LibraryService) GetArtistsSortByName(user users.User, count int, offset int) ([]model.LibraryArtist, error) {
	return svc.artists.GetArtistsSortName(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumById(user users.User, id string) (model.LibraryAlbum, error) {
	album, err := svc.albums.GetAlbumById(user.Id, id)
	if err != nil {
		return model.LibraryAlbum{}, err
	}

	tracks, err := svc.tracks.GetTracksByAlbumId(user.Id, album.Id)
	if err != nil {
		return model.LibraryAlbum{}, err
	}

	album.Tracks = tracks

	return album, err
}

func (svc *LibraryService) SearchAlbumsByQuery(user users.User, query string, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.SearchAlbumsByQuery(user.Id, query, count, offset)
}

func (svc *LibraryService) GetAlbumsSortById(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortId(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumsSortByAddedAtDesc(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortAddedAtDesc(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumsSortByReleasedAtDesc(user users.User, count int, offset int, fromYear int, toYear int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortReleasedAtDesc(user.Id, count, offset, fromYear, toYear)
}

func (svc *LibraryService) GetAlbumsSortByPlayedAtDesc(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortPlayedAtDesc(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumsSortByTotalListenTimeDesc(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortTotalListenedDesc(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumsSortByTitle(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortTitle(user.Id, count, offset)
}

func (svc *LibraryService) GetAlbumsSortByArtist(user users.User, count int, offset int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortArtist(user.Id, count, offset)
}

func (svc *LibraryService) GetRandomAlbums(user users.User, count int) ([]model.LibraryAlbum, error) {
	return svc.albums.GetAlbumsSortRandom(user.Id, count)
}

func (svc *LibraryService) GetTrackById(userId uuid.UUID, id string) (model.LibraryTrack, error) {
	return svc.tracks.GetTrackById(userId, id)
}

func (svc *LibraryService) GetTracksByIds(userId uuid.UUID, ids []string) ([]model.LibraryTrack, error) {
	return svc.tracks.GetTracksByIds(userId, ids)
}

func (svc *LibraryService) SearchTracksByQuery(userId uuid.UUID, query string, count int, offset int) ([]model.LibraryTrack, error) {
	return svc.tracks.SearchTracksByQuery(userId, query, count, offset)
}

func (svc *LibraryService) SearchTracksByFields(userId uuid.UUID, filter TrackFilter, count int, offset int) ([]model.LibraryTrack, error) {
	return svc.tracks.SearchTracksByFields(userId, filter, count, offset)
}

func (svc *LibraryService) GetTracksSortById(user users.User, count int, offset int) ([]model.LibraryTrack, error) {
	return svc.tracks.GetTracksSortId(user.Id, count, offset)
}

func (svc *LibraryService) GetRandomTracks(user users.User, count int, fromYear *int, toYear *int) ([]model.LibraryTrack, error) {
	return svc.tracks.GetTracksSortRandom(user.Id, count, fromYear, toYear)
}

func (svc *LibraryService) GetPlaylistById(user users.User, id string) (model.LibraryPlaylist, error) {
	playlist, err := svc.playlists.GetPlaylistById(user.Id, id)
	if err != nil {
		return model.LibraryPlaylist{}, err
	}

	trackIds, err := svc.playlists.GetTrackIdsByPlaylistId(playlist.Id)
	if err != nil {
		return model.LibraryPlaylist{}, err
	}

	stringIds := []string{}
	ordering := map[string]int{}
	for index, trackId := range trackIds {
		stringIds = append(stringIds, trackId.String())
		ordering[trackId.String()] = index
	}

	playlist.Tracks, err = svc.GetTracksByIds(user.Id, stringIds)
	if err != nil {
		return model.LibraryPlaylist{}, err
	}

	slices.SortFunc(playlist.Tracks, func(a model.LibraryTrack, b model.LibraryTrack) int { return ordering[a.Id] - ordering[b.Id] })

	return playlist, err
}

func (svc *LibraryService) GetPlaylists(user users.User) ([]model.LibraryPlaylist, error) {
	return svc.playlists.GetAllPlaylists(user.Id)
}
