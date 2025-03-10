package tasks

import (
	"fmt"
	"log/slog"
	"strings"
	"tapesonic/http/listenbrainz"
	"tapesonic/logic"
	"tapesonic/storage"
)

const providerListenbrainz = "listenbrainz"

type ListenBrainzPlaylistSyncHandler struct {
	client      *listenbrainz.ListenBrainzClient
	cachedSongs *logic.SongCacheService
	playlists   *storage.ExternalPlaylistStorage
}

func NewListenBrainzPlaylistSyncHandler(
	client *listenbrainz.ListenBrainzClient,
	cachedSongs *logic.SongCacheService,
	playlists *storage.ExternalPlaylistStorage,
) *ListenBrainzPlaylistSyncHandler {
	return &ListenBrainzPlaylistSyncHandler{
		client:      client,
		cachedSongs: cachedSongs,
		playlists:   playlists,
	}
}

func (h *ListenBrainzPlaylistSyncHandler) Name() string {
	return "LISTENBRAINZ_PLAYLIST_SYNC"
}

func (h *ListenBrainzPlaylistSyncHandler) OnSchedule() error {
	slog.Debug("Synchronizing ListenBrainz playlists")

	tokenInfo, err := h.client.ValidateToken()
	if err != nil || !tokenInfo.Valid {
		return fmt.Errorf("failed to get ListenBrainz username: %w", err)
	}

	slog.Debug(fmt.Sprintf("Synchronizing last.fm playlists for %s", tokenInfo.Username))

	playlists, err := h.client.GetPlaylistsCreatedFor(tokenInfo.Username, 20, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch \"Created for you\" playlists from ListenBrainz: %w", err)
	}

	resultPlaylists := []storage.ExternalPlaylist{}
	for _, playlist := range playlists.Playlists {
		resultPlaylist, err := h.processPlaylist(playlist.Playlist)
		if err != nil {
			return fmt.Errorf("failed to process ListenBrainz playlist %s: %w", playlist.Playlist.Title, err)
		}

		resultPlaylists = append(resultPlaylists, resultPlaylist)
	}

	err = h.playlists.Replace(providerListenbrainz, resultPlaylists)
	if err != nil {
		return fmt.Errorf("failed to replace ListenBrainz playlists in the database: %w", err)
	}

	slog.Info("Done synchronizing ListenBrainz playlists")
	return nil
}

func (h *ListenBrainzPlaylistSyncHandler) processPlaylist(playlist listenbrainz.PlaylistResponse) (storage.ExternalPlaylist, error) {
	slog.Debug(fmt.Sprintf("Processing ListenBrainz playlist %s", playlist.Title))

	playlistIdParts := strings.Split(playlist.Identifier, "/")
	if len(playlistIdParts) == 0 {
		return storage.ExternalPlaylist{}, fmt.Errorf("unable to parse id from `%s`", playlist.Identifier)
	}

	playlistId := playlistIdParts[len(playlistIdParts)-1]

	playlistInfo, err := h.client.GetPlaylist(playlistId)
	if err != nil {
		return storage.ExternalPlaylist{}, err
	}

	resultPlaylist := storage.ExternalPlaylist{
		Id:          fmt.Sprintf("%s_%s", providerListenbrainz, playlistId),
		Provider:    providerListenbrainz,
		RawId:       playlistId,
		Name:        playlistInfo.Title,
		Description: playlistInfo.Annotation,
		CreatedBy:   playlistInfo.Creator,
		CreatedAt:   playlist.Date,
	}

	for i, track := range playlistInfo.Track {
		targetTrackText := fmt.Sprintf("artist=%s, album=%s, title=%s", track.Creator, track.Album, track.Title)

		libraryTrack, err := h.cachedSongs.FindCachedSongByFields(track.Creator, track.Title, track.Album)
		if err != nil {
			return storage.ExternalPlaylist{}, fmt.Errorf("failed to search for a library track: %w", err)
		}

		if libraryTrack == nil {
			slog.Debug(fmt.Sprintf("Didn't find track [%s] in library, retrying search without album", targetTrackText))

			libraryTrack, err = h.cachedSongs.FindCachedSongByFields(track.Creator, track.Title, "")
			if err != nil {
				return storage.ExternalPlaylist{}, fmt.Errorf("failed to search for a library track: %w", err)
			}

			if libraryTrack == nil {
				slog.Debug(fmt.Sprintf("Didn't find track [%s] in library even without album, skipping", targetTrackText))
				continue
			}
		}

		libraryTrackText := fmt.Sprintf("artist=%s, album=%s, title=%s", libraryTrack.Artist, libraryTrack.Album, libraryTrack.Title)
		slog.Debug(fmt.Sprintf("Found track [%s] in library: %s %s [%s]", targetTrackText, libraryTrack.ServiceName, libraryTrack.SongId, libraryTrackText))

		resultTrack := storage.ExternalPlaylistTrack{
			Artist: track.Creator,
			Album:  track.Album,
			Title:  track.Title,

			ExternalPlaylist: &resultPlaylist,

			MatchedServiceName: libraryTrack.ServiceName,
			MatchedSongId:      libraryTrack.SongId,

			TrackIndex: i,
		}

		resultPlaylist.Tracks = append(resultPlaylist.Tracks, resultTrack)
	}

	slog.Debug(
		fmt.Sprintf(
			"Found %d/%d tracks for the ListenBrainz playlist `%s` id=%s",
			len(resultPlaylist.Tracks),
			len(playlistInfo.Track),
			playlistInfo.Title,
			playlistId,
		),
	)

	return resultPlaylist, nil
}
