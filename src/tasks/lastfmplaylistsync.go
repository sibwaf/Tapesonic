package tasks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"tapesonic/http/lastfm"
	"tapesonic/logic"
	"tapesonic/storage"
)

const providerLastfm = "lastfm"

type LastFmPlaylistSyncHandler struct {
	client      *lastfm.LastFmClient
	svc         *logic.LastFmService
	cachedSongs *logic.SongCacheService
	importer    *logic.AutoImportService
	playlists   *storage.ExternalPlaylistStorage

	targetPlaylistSize int
}

func NewLastFmPlaylistSyncHandler(
	client *lastfm.LastFmClient,
	svc *logic.LastFmService,
	cachedSongs *logic.SongCacheService,
	importer *logic.AutoImportService,
	playlists *storage.ExternalPlaylistStorage,
	targetPlaylistSize int,
) *LastFmPlaylistSyncHandler {
	return &LastFmPlaylistSyncHandler{
		client:             client,
		svc:                svc,
		cachedSongs:        cachedSongs,
		importer:           importer,
		playlists:          playlists,
		targetPlaylistSize: targetPlaylistSize,
	}
}

func (h *LastFmPlaylistSyncHandler) Name() string {
	return "LASTFM_PLAYLIST_SYNC"
}

func (h *LastFmPlaylistSyncHandler) OnSchedule() error {
	slog.Debug("Synchronizing last.fm playlists")

	session, err := h.svc.GetCurrentSession()
	if err != nil {
		return fmt.Errorf("failed to get current last.fm session: %w", err)
	}

	if session == nil {
		slog.Debug("No current last.fm session found, skipping playlist sync")
		return nil
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
		return fmt.Errorf("failed to get library playlist for %s from last.fm: %w", session.Username, err)
	}

	mixPlaylist, err := h.processPlaylist(
		fmt.Sprintf("%s_mix", session.Username),
		"last.fm: Mix",
		func(page int) (lastfm.PlaylistWrapper, error) {
			return h.client.GetMixPlaylist(session.Username, page)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get mix playlist for %s from last.fm: %w", session.Username, err)
	}

	recommendedPlaylist, err := h.processPlaylist(
		fmt.Sprintf("%s_recommended", session.Username),
		"last.fm: Recommended",
		func(page int) (lastfm.PlaylistWrapper, error) {
			return h.client.GetRecommendedPlaylist(session.Username, page)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get recommended playlist for %s from last.fm: %w", session.Username, err)
	}

	err = h.playlists.Replace(providerLastfm, []storage.ExternalPlaylist{libraryPlaylist, mixPlaylist, recommendedPlaylist})
	if err != nil {
		return fmt.Errorf("failed to replace last.fm playlists in the database: %w", err)
	}

	slog.Info("Done synchronizing last.fm playlists")
	return nil
}

func (h *LastFmPlaylistSyncHandler) processPlaylist(
	playlistId string,
	playlistName string,
	fetch func(page int) (lastfm.PlaylistWrapper, error),
) (storage.ExternalPlaylist, error) {
	slog.Debug(fmt.Sprintf("Synchronizing last.fm playlist %s", playlistId))

	type trackOrErr struct {
		track lastfm.PlaylistItem
		err   error
	}

	rawTracks := make(chan trackOrErr)
	producerCtx, cancelProducer := context.WithCancel(context.Background())

	go func() {
		page := 1
		tracksSeen := 0
		for {
			if tracksSeen >= h.targetPlaylistSize*3 {
				slog.Warn(fmt.Sprintf("Processed %d raw tracks while trying to collect %d tracks for last.fm playlist %s, giving up", tracksSeen, h.targetPlaylistSize, playlistId))
				close(rawTracks)
				break
			}

			playlist, err := fetch(page)
			if err != nil {
				rawTracks <- trackOrErr{err: fmt.Errorf("failed to fetch playlist page: %w", err)}
				close(rawTracks)
				break
			}

			slog.Debug(fmt.Sprintf("Fetched %d more tracks from last.fm playlist %s page %d", len(playlist.Items), playlistId, page))

			if len(playlist.Items) == 0 {
				slog.Warn(fmt.Sprintf("last.fm playlist %s has no tracks on page %d, stopping enumeration", playlistId, page))
				close(rawTracks)
				break
			}

			for _, track := range playlist.Items {
				select {
				case <-producerCtx.Done():
					return
				case rawTracks <- trackOrErr{track: track}:
					tracksSeen += 1
				}
			}

			page += 1
		}
	}()

	defer cancelProducer()

	trackIds := map[string]bool{}
	tracks := []storage.ExternalPlaylistTrack{}
	for trackOrErr := range rawTracks {
		if len(tracks) >= h.targetPlaylistSize {
			break
		}

		if trackOrErr.err != nil {
			return storage.ExternalPlaylist{}, trackOrErr.err
		}

		track := trackOrErr.track

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

		var libraryTrackText string
		if libraryTrack != nil {
			libraryTrackText = fmt.Sprintf("service=%s, id=%s, artist=%s, title=%s", libraryTrack.ServiceName, libraryTrack.SongId, libraryTrack.Artist, libraryTrack.Title)
			slog.Debug(fmt.Sprintf("Found track [%s] in library: [%s]", targetTrackText, libraryTrackText))
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

			libraryTrack = &cachedTrack
			libraryTrackText = fmt.Sprintf("service=%s, id=%s, artist=%s, title=%s", libraryTrack.ServiceName, libraryTrack.SongId, libraryTrack.Artist, libraryTrack.Title)
			slog.Debug(fmt.Sprintf("Imported track [%s] from %s for playlist %s: [%s]", targetTrackText, url, playlistId, libraryTrackText))
		}

		deduplicationId := fmt.Sprintf("%s/%s", libraryTrack.ServiceName, libraryTrack.SongId)
		if _, alreadyAdded := trackIds[deduplicationId]; alreadyAdded {
			slog.Debug(fmt.Sprintf("Track [%s] was already added to playlist %s, skipping", libraryTrackText, playlistId))
			continue
		} else {
			trackIds[deduplicationId] = true
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
