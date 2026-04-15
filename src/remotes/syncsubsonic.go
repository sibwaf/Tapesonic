package remotes

import (
	"fmt"
	"log/slog"
	"tapesonic/artists"
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

// todo: logging
// todo: playlists

const (
	batchSize = 100
)

type SubsonicSyncService struct {
	allArtists *artists.ArtistService

	remotes  *RemoteStorage
	artworks *RemoteArtworkStorage
	artists  *RemoteArtistStorage
	albums   *RemoteAlbumStorage
	tracks   *RemoteTrackStorage
}

func newSubsonicSyncService(
	allArtists *artists.ArtistService,
	remotes *RemoteStorage,
	artworks *RemoteArtworkStorage,
	artists *RemoteArtistStorage,
	albums *RemoteAlbumStorage,
	tracks *RemoteTrackStorage,
) *SubsonicSyncService {
	return &SubsonicSyncService{
		allArtists: allArtists,
		remotes:    remotes,
		artworks:   artworks,
		artists:    artists,
		albums:     albums,
		tracks:     tracks,
	}
}

func (svc *SubsonicSyncService) SyncLibrary(userId uuid.UUID, remoteId uuid.UUID) error {
	slog.Info(fmt.Sprintf("Syncing Subsonic library remote id=%s for user id=%s", remoteId, userId))

	remote, err := svc.remotes.FindById(remoteId)
	if err != nil {
		return err
	}
	if remote == nil {
		return fmt.Errorf("remote doesn't exist: %w", model.ErrNotFound)
	}

	credentials, err := svc.remotes.FindCredentials(userId, remote.Id)
	if err != nil {
		return err
	}
	if credentials == nil {
		return fmt.Errorf("user is not authenticated for remote: %w", model.ErrNotAuthenticated)
	}

	client := subsonic.NewSubsonicClient(remote.Url)
	auth := GetSubsonicAuthMethod(credentials)

	syncTag := uuid.New().String()

	offset := 0
	requestMoreArtists := 1
	requestMoreAlbums := 1
	requestMoreSongs := 1

	for requestMoreArtists > 0 || requestMoreAlbums > 0 || requestMoreSongs > 0 {
		response, err := client.Search3(
			auth,
			"",
			batchSize*requestMoreArtists,
			offset,
			batchSize*requestMoreAlbums,
			offset,
			batchSize*requestMoreSongs,
			offset,
		)
		if err != nil {
			return err
		}

		for _, artist := range response.Artist {
			if svc.upsertArtist(artist, remote.Id, userId, syncTag) != nil {
				return err
			}
		}
		if len(response.Artist) < batchSize {
			requestMoreArtists = 0
		}

		for _, album := range response.Album {
			if svc.upsertAlbum(album, remote.Id, userId, syncTag) != nil {
				return err
			}
		}
		if len(response.Album) < batchSize {
			requestMoreAlbums = 0
		}

		for _, song := range response.Song {
			if svc.upsertTrack(song, remote.Id, userId, syncTag) != nil {
				return err
			}
		}
		if len(response.Song) < batchSize {
			requestMoreSongs = 0
		}

		offset += batchSize
	}

	if err := svc.tracks.DeleteUserBindingsBySyncTag(userId, syncTag); err != nil {
		slog.Warn(fmt.Sprintf("Failed to cleanup remote tracks while syncing Subsonic library remote id=%s for user id=%s, latest syncTag=%s", remote.Id, userId, syncTag))
	}
	if err := svc.albums.DeleteUserBindingsBySyncTag(userId, syncTag); err != nil {
		slog.Warn(fmt.Sprintf("Failed to cleanup remote albums while syncing Subsonic library remote id=%s for user id=%s, latest syncTag=%s", remote.Id, userId, syncTag))
	}
	if err := svc.artists.DeleteUserBindingsBySyncTag(userId, syncTag); err != nil {
		slog.Warn(fmt.Sprintf("Failed to cleanup remote artists while syncing Subsonic library remote id=%s for user id=%s, latest syncTag=%s", remote.Id, userId, syncTag))
	}

	return nil
}

func (svc *SubsonicSyncService) upsertArtist(artist subsonic.ArtistId3, remoteId uuid.UUID, userId uuid.UUID, syncTag string) error {
	if artist.CoverArt != "" {
		if err := svc.artworks.Upsert(RemoteArtwork{Id: uuid.New(), RemoteId: remoteId, ArtworkId: artist.CoverArt}); err != nil {
			return err
		}
	}

	upsertedArtist, err := svc.artists.Upsert(
		RemoteArtist{
			Id:            uuid.New(),
			RemoteId:      remoteId,
			ArtistId:      artist.Id,
			ArtworkId:     artist.CoverArt,
			Name:          artist.Name,
			MusicBrainzId: artist.MusicBrainzId,
		},
		RemoteArtistToUser{
			UserId:  userId,
			SyncTag: syncTag,
		},
	)
	if err != nil {
		return err
	}

	if upsertedArtist.TapesonicArtistId == nil {
		libraryArtist, err := svc.allArtists.FindMatchOrCreate(upsertedArtist.Name, upsertedArtist.MusicBrainzId)
		if err != nil {
			return err
		}

		err = svc.artists.LinkToTapesonicArtist(upsertedArtist.Id, libraryArtist.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *SubsonicSyncService) upsertAlbum(album subsonic.AlbumId3, remoteId uuid.UUID, userId uuid.UUID, syncTag string) error {
	if album.CoverArt != "" {
		if err := svc.artworks.Upsert(RemoteArtwork{Id: uuid.New(), RemoteId: remoteId, ArtworkId: album.CoverArt}); err != nil {
			return err
		}
	}

	releasedAt := time.Time{}
	if album.ReleaseDate != nil {
		releasedAt = time.Date(album.ReleaseDate.Year, time.Month(album.ReleaseDate.Month), album.ReleaseDate.Day, 0, 0, 0, 0, time.UTC)
	} else if album.Year != 0 {
		releasedAt = time.Date(album.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return svc.albums.Upsert(
		RemoteAlbum{
			Id:         uuid.New(),
			RemoteId:   remoteId,
			AlbumId:    album.Id,
			ArtworkId:  album.CoverArt,
			ArtistId:   album.ArtistId,
			Title:      album.Name,
			AddedAt:    util.NewTimestampWrapper(album.Created),
			ReleasedAt: util.NewTimestampWrapperOrNull(&releasedAt),
		},
		RemoteAlbumToUser{
			UserId:  userId,
			SyncTag: syncTag,
		},
	)
}

func (svc *SubsonicSyncService) upsertTrack(song subsonic.Child, remoteId uuid.UUID, userId uuid.UUID, syncTag string) error {
	if song.CoverArt != "" {
		if err := svc.artworks.Upsert(RemoteArtwork{Id: uuid.New(), RemoteId: remoteId, ArtworkId: song.CoverArt}); err != nil {
			return err
		}
	}

	return svc.tracks.Upsert(
		RemoteTrack{
			Id:         uuid.New(),
			RemoteId:   remoteId,
			TrackId:    song.Id,
			Artist:     song.Artist,
			ArtistId:   song.ArtistId,
			Album:      song.Album,
			AlbumId:    song.AlbumId,
			AlbumIndex: song.Track - 1,
			ArtworkId:  song.CoverArt,
			Title:      song.Title,
			DurationMs: song.Duration * 1000,
		},
		RemoteTrackToUser{
			UserId:  userId,
			SyncTag: syncTag,
		},
	)
}
