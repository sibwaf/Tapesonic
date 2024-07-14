package logic

import (
	"tapesonic/http/listenbrainz"
	"time"
)

type ScrobbleService struct {
	listenbrainz *listenbrainz.ListenBrainzClient
}

func NewScrobbleService(listenbrainz listenbrainz.ListenBrainzClient) *ScrobbleService {
	return &ScrobbleService{
		listenbrainz: &listenbrainz,
	}
}

func (svc *ScrobbleService) ScrobblePlaying(
	artist string,
	album string,
	track string,
) error {
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

	return svc.listenbrainz.SubmitListens(request)
}

func (svc *ScrobbleService) ScrobbleCompleted(
	listenedAt time.Time,
	artist string,
	album string,
	track string,
) error {
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

	return svc.listenbrainz.SubmitListens(request)
}
