package scrobbling

import (
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/lastfm"
	"tapesonic/library"
	"tapesonic/listenbrainz"
	"tapesonic/model"
	"tapesonic/remotes"
	"tapesonic/subsonic"
	"tapesonic/users"
	"time"

	"github.com/google/uuid"
)

type ScrobbleService struct {
	listenbrainz *listenbrainz.ListenBrainzService
	lastfm       *lastfm.LastFmService
	library      *library.LibraryService
	remotes      *remotes.RemoteService
	stats        *ListenStatStorage
}

func NewScrobbleService(
	listenbrainz *listenbrainz.ListenBrainzService,
	lastfm *lastfm.LastFmService,
	library *library.LibraryService,
	remotes *remotes.RemoteService,
	stats *ListenStatStorage,
) *ScrobbleService {
	return &ScrobbleService{
		listenbrainz: listenbrainz,
		lastfm:       lastfm,
		library:      library,
		remotes:      remotes,
		stats:        stats,
	}
}

func (svc *ScrobbleService) ScrobblePlaying(
	user users.User,
	trackId string,
	listenedAt time.Time,
) error {
	track, err := svc.library.GetTrackById(user.Id, trackId)
	if err != nil {
		return err
	}

	remote := remotes.Remote{}
	if track.RemoteId != nil {
		remote, err = svc.remotes.GetById(*track.RemoteId)
		if err != nil {
			return err
		}
	}

	err = svc.stats.TouchListenedAt(user.Id, trackId, listenedAt)
	if err != nil {
		return err
	}

	scrobbleErrors := []error{}

	if remote.Id != uuid.Nil && remote.IsScrobbleReplicationEnabled {
		if track.RemoteTrackId == nil {
			scrobbleErrors = append(scrobbleErrors, fmt.Errorf("track id=%s is remote, but is missing remote_track_id", track.Id))
		} else {
			scrobbleErrors = append(scrobbleErrors, svc.scrobblePlayingToRemote(remote, user, *track.RemoteTrackId, listenedAt))
		}
	}

	if remote.Id == uuid.Nil || remote.IsExternalScrobblingEnabled {
		scrobbleErrors = append(scrobbleErrors, svc.scrobblePlayingToExternalServices(user, track))
	}

	return errors.Join(scrobbleErrors...)
}

func (svc *ScrobbleService) scrobblePlayingToRemote(remote remotes.Remote, user users.User, remoteTrackId string, listenedAt time.Time) error {
	credentials, err := svc.remotes.GetCredentials(user, remote)
	if errors.Is(err, model.ErrNotAuthenticated) {
		slog.Warn(fmt.Sprintf("Skipping \"playing now\" scrobble remoteId=%s, remoteTrackId=%s because user is not authenticated in this remote", remote.Id, remoteTrackId))
		return nil
	}

	switch remote.Type {
	case remotes.REMOTE_TYPE_SUBSONIC:
		auth := remotes.GetSubsonicAuthMethod(&credentials)
		client := subsonic.NewSubsonicClient(remote.Url)
		return client.Scrobble(auth, remoteTrackId, listenedAt, false)
	default:
		slog.Error(fmt.Sprintf("Skipping \"playing now\" scrobble to remote remoteId=%s: unknown remote type %s", remote.Id, remote.Type))
		return nil
	}
}

func (svc *ScrobbleService) scrobblePlayingToExternalServices(user users.User, track model.LibraryTrack) error {
	if track.ArtistName == "" || track.Title == "" {
		slog.Debug(fmt.Sprintf("Skipping \"playing now\" scrobble to external services because artist or track is missing: artist=%s, track=%s, album=%s", track.ArtistName, track.Title, track.AlbumName))
		return nil
	}

	lastFmErr := svc.lastfm.UpdateNowPlaying(user, track.ArtistName, track.Title, track.AlbumName)
	if errors.Is(lastFmErr, lastfm.ErrLastFmNotConfigured) || errors.Is(lastFmErr, lastfm.ErrNoActiveSession) {
		lastFmErr = nil
	}

	listenbrainzErr := svc.listenbrainz.UpdateNowPlaying(user, track.ArtistName, track.Title, track.AlbumName)
	if errors.Is(listenbrainzErr, listenbrainz.ErrNoActiveSession) {
		listenbrainzErr = nil
	}

	return errors.Join(lastFmErr, listenbrainzErr)
}

func (svc *ScrobbleService) ScrobbleCompleted(
	user users.User,
	trackId string,
	listenedAt time.Time,
) error {
	// todo: outbox

	track, err := svc.library.GetTrackById(user.Id, trackId)
	if err != nil {
		return err
	}

	remote := remotes.Remote{}
	if track.RemoteId != nil {
		remote, err = svc.remotes.GetById(*track.RemoteId)
		if err != nil {
			return err
		}
	}

	err = svc.stats.AddListen(user.Id, trackId, listenedAt)
	if err != nil {
		return err
	}

	scrobbleErrors := []error{}

	if remote.Id != uuid.Nil && remote.IsScrobbleReplicationEnabled {
		if track.RemoteTrackId == nil {
			scrobbleErrors = append(scrobbleErrors, fmt.Errorf("track id=%s is remote, but is missing remote_track_id", track.Id))
		} else {
			scrobbleErrors = append(scrobbleErrors, svc.scrobbleCompletedToRemote(remote, user, *track.RemoteTrackId, listenedAt))
		}
	}

	if remote.Id == uuid.Nil || remote.IsExternalScrobblingEnabled {
		scrobbleErrors = append(scrobbleErrors, svc.scrobbleCompletedToExternalServices(user, track, listenedAt))
	}

	return errors.Join(scrobbleErrors...)
}

func (svc *ScrobbleService) scrobbleCompletedToRemote(remote remotes.Remote, user users.User, remoteTrackId string, listenedAt time.Time) error {
	credentials, err := svc.remotes.GetCredentials(user, remote)
	if errors.Is(err, model.ErrNotAuthenticated) {
		slog.Warn(fmt.Sprintf("Skipping \"completed\" scrobble remoteId=%s, remoteTrackId=%s because user is not authenticated in this remote", remote.Id, remoteTrackId))
		return nil
	}

	switch remote.Type {
	case remotes.REMOTE_TYPE_SUBSONIC:
		auth := remotes.GetSubsonicAuthMethod(&credentials)
		client := subsonic.NewSubsonicClient(remote.Url)
		return client.Scrobble(auth, remoteTrackId, listenedAt, true)
	default:
		slog.Error(fmt.Sprintf("Skipping \"completed\" scrobble to remote remoteId=%s: unknown remote type %s", remote.Id, remote.Type))
		return nil
	}
}

func (svc *ScrobbleService) scrobbleCompletedToExternalServices(user users.User, track model.LibraryTrack, listenedAt time.Time) error {
	if track.ArtistName == "" || track.Title == "" {
		slog.Debug(fmt.Sprintf("Skipping \"completed\" scrobble to external services because artist or track is missing: artist=%s, track=%s, album=%s", track.ArtistName, track.Title, track.AlbumName))
		return nil
	}

	lastFmErr := svc.lastfm.Scrobble(user, listenedAt, track.ArtistName, track.Title, track.AlbumName)
	if errors.Is(lastFmErr, lastfm.ErrLastFmNotConfigured) || errors.Is(lastFmErr, lastfm.ErrNoActiveSession) {
		lastFmErr = nil
	}

	listenbrainzErr := svc.listenbrainz.Scrobble(user, listenedAt, track.ArtistName, track.Title, track.AlbumName)
	if errors.Is(listenbrainzErr, listenbrainz.ErrNoActiveSession) {
		listenbrainzErr = nil
	}

	return errors.Join(lastFmErr, listenbrainzErr)
}
