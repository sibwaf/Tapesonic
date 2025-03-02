package tasks

import (
	"fmt"
	"log/slog"
	"strings"
	"tapesonic/config"
	"tapesonic/http/listenbrainz"
	"tapesonic/logic"
	"tapesonic/storage"

	"github.com/robfig/cron/v3"
)

const providerListenbrainz = "listenbrainz"

type ListenBrainzPlaylistSyncHandler struct {
	client      *listenbrainz.ListenBrainzClient
	cachedSongs *logic.SongCacheService
	playlists   *storage.ExternalPlaylistStorage

	taskConfig config.BackgroundTaskConfig
}

func NewListenBrainzPlaylistSyncHandler(
	client *listenbrainz.ListenBrainzClient,
	cachedSongs *logic.SongCacheService,
	playlists *storage.ExternalPlaylistStorage,

	taskConfig config.BackgroundTaskConfig,
) *ListenBrainzPlaylistSyncHandler {
	return &ListenBrainzPlaylistSyncHandler{
		client:      client,
		cachedSongs: cachedSongs,
		playlists:   playlists,

		taskConfig: taskConfig,
	}
}

func (h *ListenBrainzPlaylistSyncHandler) RegisterSchedules(cron *cron.Cron) error {
	_, err := cron.AddFunc(h.taskConfig.Cron, h.onSchedule)
	return err
}

func (h *ListenBrainzPlaylistSyncHandler) onSchedule() {
	slog.Debug("Synchronizing ListenBrainz playlists")

	tokenInfo, err := h.client.ValidateToken()
	if err != nil || !tokenInfo.Valid {
		slog.Error(fmt.Sprintf("Failed to get ListenBrainz username: %s", err.Error()))
		return
	}

	playlists, err := h.client.GetPlaylistsCreatedFor(tokenInfo.Username, 20, 0)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to fetch \"Created for you\" playlists from ListenBrainz: %s", err.Error()))
		return
	}

	resultPlaylists := []storage.ExternalPlaylist{}
	for _, playlist := range playlists.Playlists {
		resultPlaylist, err := h.processPlaylist(playlist.Playlist)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to process ListenBrainz playlist `%s`: %s", playlist.Playlist.Title, err.Error()))
			return
		}

		resultPlaylists = append(resultPlaylists, resultPlaylist)
	}

	err = h.playlists.Replace(providerListenbrainz, resultPlaylists)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to replace ListenBrainz playlists in the database: %s", err.Error()))
	}

	slog.Info("Done synchronizing ListenBrainz playlists")
}

func (h *ListenBrainzPlaylistSyncHandler) processPlaylist(playlist listenbrainz.PlaylistResponse) (storage.ExternalPlaylist, error) {
	slog.Debug(fmt.Sprintf("Processing ListenBrainz playlist `%s`", playlist.Title))

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
		trackId, err := h.cachedSongs.FindSongIdByFields(track.Creator, track.Title, track.Album)
		if err != nil {
			return storage.ExternalPlaylist{}, err
		}

		if trackId == nil {
			slog.Debug(fmt.Sprintf("Didn't manage to find a matching track in the library: artist=`%s`, album=`%s`, title=`%s`", track.Creator, track.Album, track.Title))
			continue
		}

		resultTrack := storage.ExternalPlaylistTrack{
			Artist: track.Creator,
			Album:  track.Album,
			Title:  track.Title,

			ExternalPlaylist: &resultPlaylist,

			MatchedServiceName: trackId.ServiceName,
			MatchedSongId:      trackId.Id,

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
