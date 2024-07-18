package tasks

import (
	"fmt"
	"log/slog"
	"strings"
	"tapesonic/config"
	"tapesonic/http/listenbrainz"
	"tapesonic/storage"

	"github.com/robfig/cron/v3"
)

type ListenBrainzPlaylistSyncHandler struct {
	client                *listenbrainz.ListenBrainzClient
	songs                 *storage.CachedMuxSongStorage
	listenbrainzPlaylists storage.ListenbrainzPlaylistStorage

	taskConfig config.BackgroundTaskConfig
}

func NewListenBrainzPlaylistSyncHandler(
	client *listenbrainz.ListenBrainzClient,
	songs *storage.CachedMuxSongStorage,
	listenbrainzPlaylists storage.ListenbrainzPlaylistStorage,

	taskConfig config.BackgroundTaskConfig,
) *ListenBrainzPlaylistSyncHandler {
	return &ListenBrainzPlaylistSyncHandler{
		client:                client,
		songs:                 songs,
		listenbrainzPlaylists: listenbrainzPlaylists,

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

	resultPlaylists := []storage.ListenbrainzPlaylist{}
	for _, playlist := range playlists.Playlists {
		resultPlaylist, err := h.processPlaylist(playlist.Playlist)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to process ListenBrainz playlist `%s`: %s", playlist.Playlist.Title, err.Error()))
			return
		}

		resultPlaylists = append(resultPlaylists, resultPlaylist)
	}

	err = h.listenbrainzPlaylists.Replace(resultPlaylists)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to replace ListenBrainz playlists in the database: %s", err.Error()))
	}

	slog.Info("Done synchronizing ListenBrainz playlists")
}

func (h *ListenBrainzPlaylistSyncHandler) processPlaylist(playlist listenbrainz.PlaylistResponse) (storage.ListenbrainzPlaylist, error) {
	slog.Debug(fmt.Sprintf("Processing ListenBrainz playlist `%s`", playlist.Title))

	playlistIdParts := strings.Split(playlist.Identifier, "/")
	if len(playlistIdParts) == 0 {
		return storage.ListenbrainzPlaylist{}, fmt.Errorf("unable to parse id from `%s`", playlist.Identifier)
	}

	playlistId := playlistIdParts[len(playlistIdParts)-1]

	playlistInfo, err := h.client.GetPlaylist(playlistId)
	if err != nil {
		return storage.ListenbrainzPlaylist{}, err
	}

	resultPlaylist := storage.ListenbrainzPlaylist{
		Id:          playlistId,
		Name:        playlistInfo.Title,
		Description: playlistInfo.Annotation,
		CreatedBy:   playlistInfo.Creator,
		CreatedAt:   playlist.Date,
	}

	for i, track := range playlistInfo.Track {
		trackId, err := h.findTrack(track.Creator, track.Album, track.Title)
		if err != nil {
			return storage.ListenbrainzPlaylist{}, err
		}

		if trackId == nil {
			slog.Debug(fmt.Sprintf("Didn't manage to find a matching track in the library: artist=`%s`, album=`%s`, title=`%s`", track.Creator, track.Album, track.Title))
			continue
		}

		resultTrack := storage.ListenbrainzPlaylistTrack{
			Artist: track.Creator,
			Album:  track.Album,
			Title:  track.Title,

			ListenbrainzPlaylist: &resultPlaylist,

			MatchedServiceName: trackId.ServiceName,
			MatchedSongId:      trackId.Id,

			TrackIndex: i,
		}

		resultPlaylist.Tracks = append(resultPlaylist.Tracks, &resultTrack)
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

func (h *ListenBrainzPlaylistSyncHandler) findTrack(artist string, album string, title string) (*storage.CachedSongId, error) {
	tracks, err := h.songs.SearchByFields(artist, album, title, 2)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 1 {
		return &tracks[0], nil
	}

	return nil, nil
}
