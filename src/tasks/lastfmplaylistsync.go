package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"tapesonic/config"
	"tapesonic/http/lastfm"
	"tapesonic/logic"
	"tapesonic/storage"

	"github.com/robfig/cron/v3"
)

const providerLastfm = "lastfm"

type LastFmPlaylistSyncHandler struct {
	client      *lastfm.LastFmClient
	svc         *logic.LastFmService
	cachedSongs *logic.SongCacheService
	importer    *logic.AutoImportService
	playlists   *storage.ExternalPlaylistStorage

	targetPlaylistSize int

	taskConfig config.BackgroundTaskConfig
}

func NewLastFmPlaylistSyncHandler(
	client *lastfm.LastFmClient,
	svc *logic.LastFmService,
	cachedSongs *logic.SongCacheService,
	importer *logic.AutoImportService,
	playlists *storage.ExternalPlaylistStorage,
	targetPlaylistSize int,
	taskConfig config.BackgroundTaskConfig,
) *LastFmPlaylistSyncHandler {
	return &LastFmPlaylistSyncHandler{
		client:             client,
		svc:                svc,
		cachedSongs:        cachedSongs,
		importer:           importer,
		playlists:          playlists,
		targetPlaylistSize: targetPlaylistSize,
		taskConfig:         taskConfig,
	}
}

func (h *LastFmPlaylistSyncHandler) RegisterSchedules(cron *cron.Cron) error {
	_, err := cron.AddFunc(h.taskConfig.Cron, h.onSchedule)
	return err
}

func (h *LastFmPlaylistSyncHandler) onSchedule() {
	slog.Debug("Synchronizing last.fm playlists")

	session, err := h.svc.GetCurrentSession()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get current last.fm session: %s", err.Error()))
		return
	}

	if session == nil {
		slog.Debug("No current last.fm session found, skipping playlist sync")
		return
	}

	slog.Debug(fmt.Sprintf("Synchronizing last.fm playlists for %s", session.Username))

	libraryPlaylist, err := h.processPlaylist(
		fmt.Sprintf("%s_library", session.Username),
		"last.fm: Library",
		func(page int) (lastfm.PlaylistWrapper, error) {
			return h.client.GetLibraryPlaylist(session.Username, page)
		},
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get library playlist for %s from last.fm: %s", session.Username, err.Error()))
		return
	}

	mixPlaylist, err := h.processPlaylist(
		fmt.Sprintf("%s_mix", session.Username),
		"last.fm: Mix",
		func(page int) (lastfm.PlaylistWrapper, error) {
			return h.client.GetMixPlaylist(session.Username, page)
		},
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get mix playlist for %s from last.fm: %s", session.Username, err.Error()))
		return
	}

	recommendedPlaylist, err := h.processPlaylist(
		fmt.Sprintf("%s_recommended", session.Username),
		"last.fm: Recommended",
		func(page int) (lastfm.PlaylistWrapper, error) {
			return h.client.GetRecommendedPlaylist(session.Username, page)
		},
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get recommended playlist for %s from last.fm: %s", session.Username, err.Error()))
		return
	}

	err = h.playlists.Replace(providerLastfm, []storage.ExternalPlaylist{libraryPlaylist, mixPlaylist, recommendedPlaylist})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to replace last.fm playlists in the database: %s", err.Error()))
	}

	slog.Info("Done synchronizing last.fm playlists")
}

func (h *LastFmPlaylistSyncHandler) processPlaylist(
	playlistId string,
	playlistName string,
	fetch func(page int) (lastfm.PlaylistWrapper, error),
) (storage.ExternalPlaylist, error) {
	slog.Debug(fmt.Sprintf("Synchronizing last.fm playlist %s", playlistId))

	sourceTracks := []lastfm.PlaylistItem{}
	page := 1
	for len(sourceTracks) < h.targetPlaylistSize {
		playlist, err := fetch(page)
		if err != nil {
			return storage.ExternalPlaylist{}, fmt.Errorf("failed to fetch playlist page: %w", err)
		}

		slog.Debug(fmt.Sprintf("Fetched %d more tracks from last.fm playlist %s page %d", len(playlist.Items), playlistId, page))
		if len(playlist.Items) == 0 {
			break
		}

		sourceTracks = append(sourceTracks, playlist.Items...)
		page += 1
	}

	tracks := []storage.ExternalPlaylistTrack{}
	for _, track := range sourceTracks {
		artists := []string{}
		for _, artist := range track.Artists {
			artists = append(artists, artist.Name)
		}

		artist := strings.Join(artists, ", ")
		title := track.Name

		targetTrackText := fmt.Sprintf("artist=%s, title=%s", artist, title)

		libraryTrack, err := h.cachedSongs.FindCachedSongByFields(artist, title, "")
		if err != nil {
			return storage.ExternalPlaylist{}, fmt.Errorf("failed to search for a library track: %w", err)
		}

		if libraryTrack != nil {
			libraryTrackText := fmt.Sprintf("artist=%s, title=%s", artist, title)
			slog.Debug(fmt.Sprintf("Found track [%s] in library: %s %s [%s]", targetTrackText, libraryTrack.ServiceName, libraryTrack.SongId, libraryTrackText))
		} else {
			if len(track.Playlinks) == 0 {
				slog.Debug(fmt.Sprintf("Didn't find track [%s] in library and no playlinks available, skipping", targetTrackText))
				continue
			}

			url := track.Playlinks[0].Url
			slog.Debug(fmt.Sprintf("Didn't find track [%s] in library, trying to import from %s", targetTrackText, url))

			importedTrack, err := h.importer.ImportTrackFrom(context.Background(), url, artist, title)
			if err != nil {
				slog.Warn(fmt.Sprintf("Failed to import track [%s] from %s, skipping: %s", targetTrackText, url, err.Error()))
				continue
			}

			// todo: refresh cache on write in TrackService/TapeService
			cachedTrack, err := h.cachedSongs.Refresh("tapesonic", importedTrack.Id.String())
			if err != nil {
				slog.Warn(fmt.Sprintf("Failed to update song cache for track id=%s, skipping: %s", importedTrack.Id, err.Error()))
				continue
			}

			slog.Debug(fmt.Sprintf("Imported track [%s] from %s for playlist %s", targetTrackText, url, playlistId))
			libraryTrack = &cachedTrack
		}

		tracks = append(tracks, storage.ExternalPlaylistTrack{
			Artist: artist,
			Title:  title,

			MatchedServiceName: libraryTrack.ServiceName,
			MatchedSongId:      libraryTrack.SongId,
		})
	}

	return storage.ExternalPlaylist{
		Id:        fmt.Sprintf("%s_%s", providerLastfm, playlistId),
		Provider:  providerLastfm,
		RawId:     playlistId,
		Name:      playlistName,
		CreatedBy: "last.fm",

		Tracks: tracks,
	}, nil
}
