package logic

import (
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/http/listenbrainz"
	"time"
)

type ScrobbleService struct {
	listenbrainz *listenbrainz.ListenBrainzClient
	lastfm       *LastFmService
}

func NewScrobbleService(
	listenbrainz *listenbrainz.ListenBrainzClient,
	lastfm *LastFmService,
) *ScrobbleService {
	return &ScrobbleService{
		listenbrainz: listenbrainz,
		lastfm:       lastfm,
	}
}

func (svc *ScrobbleService) ScrobblePlaying(
	artist string,
	album string,
	track string,
) error {
	if artist == "" || track == "" {
		slog.Debug(fmt.Sprintf("Skipping \"playing now\" scrobble because artist or track is missing: artist=%s, track=%s, album=%s", artist, track, album))
		return nil
	}

	lastFmErr := svc.lastfm.UpdateNowPlaying(artist, track, album)
	if errors.Is(lastFmErr, ErrLastFmNotConfigured) {
		lastFmErr = nil
	}

	var listenbrainzErr error = nil
	if svc.listenbrainz != nil {
		request := listenbrainz.SubmitListensRequest{
			ListenType: listenbrainz.ListenTypePlayingNow,
			Payload: []listenbrainz.SubmitListensRequestPayloadItem{
				{
					TrackMetadata: listenbrainz.SubmitListensRequestPayloadItemTrackMetadata{
						ArtistName:  artist,
						ReleaseName: album,
						TrackName:   track,
					},
				},
			},
		}
		listenbrainzErr = svc.listenbrainz.SubmitListens(request)
	}

	return errors.Join(lastFmErr, listenbrainzErr)
}

func (svc *ScrobbleService) ScrobbleCompleted(
	listenedAt time.Time,
	artist string,
	album string,
	track string,
) error {
	if artist == "" || track == "" {
		slog.Debug(fmt.Sprintf("Skipping \"completed\" scrobble because artist or track is missing: artist=%s, track=%s, album=%s", artist, track, album))
		return nil
	}

	lastFmErr := svc.lastfm.Scrobble(listenedAt, artist, track, album)
	if errors.Is(lastFmErr, ErrLastFmNotConfigured) {
		lastFmErr = nil
	}

	var listenbrainzErr error = nil
	if svc.listenbrainz != nil {
		request := listenbrainz.SubmitListensRequest{
			ListenType: listenbrainz.ListenTypeSingle,
			Payload: []listenbrainz.SubmitListensRequestPayloadItem{
				{
					ListenedAt: listenedAt.Unix(),
					TrackMetadata: listenbrainz.SubmitListensRequestPayloadItemTrackMetadata{
						ArtistName:  artist,
						ReleaseName: album,
						TrackName:   track,
					},
				},
			},
		}
		listenbrainzErr = svc.listenbrainz.SubmitListens(request)
	}

	return errors.Join(lastFmErr, listenbrainzErr)
}
