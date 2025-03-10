package logic

import (
	"context"
	"fmt"
	"tapesonic/model"
	"tapesonic/storage"
)

type AutoImportService struct {
	sources      *SourceService
	tracks       *TrackService
	trackMatcher *TrackMatcher
}

func NewAutoImportService(
	sources *SourceService,
	tracks *TrackService,
	trackMatcher *TrackMatcher,
) *AutoImportService {
	return &AutoImportService{
		sources:      sources,
		tracks:       tracks,
		trackMatcher: trackMatcher,
	}
}

func (svc *AutoImportService) ImportTrackFrom(ctx context.Context, url string, expectedArtist string, expectedTitle string) (storage.Track, error) {
	source, err := svc.sources.FindByUrl(url)
	if err != nil {
		return storage.Track{}, fmt.Errorf("failed to find source by url: %w", err)
	}

	var tracks []storage.Track
	if source != nil {
		tracks, err = svc.tracks.GetDirectTracksBySource(source.Id)
		if err != nil {
			return storage.Track{}, fmt.Errorf("failed to get tracks for source: %w", err)
		}
	} else {
		tracks = []storage.Track{}
	}

	if source == nil || len(tracks) == 0 {
		importedSource, err := svc.sources.AddSource(ctx, url, model.SOURCE_MANAGEMENT_POLICY_AUTO)
		if err != nil {
			return storage.Track{}, fmt.Errorf("failed to import source: %w", err)
		}

		tracks, err = svc.tracks.GetDirectTracksBySource(importedSource.Id)
		if err != nil {
			return storage.Track{}, fmt.Errorf("failed to get tracks for source: %w", err)
		}

		source = &importedSource
	}

	if len(tracks) == 0 {
		return storage.Track{}, fmt.Errorf("no tracks were imported from url=%s", url)
	} else if len(tracks) > 1 {
		return storage.Track{}, fmt.Errorf("multiple tracks were imported from url=%s", url)
	}

	sourceIsAutoManaged := source.ManagementPolicy == model.SOURCE_MANAGEMENT_POLICY_AUTO

	track := tracks[0]
	expectedTrack := TrackForMatching{Artist: expectedArtist, Title: expectedTitle}

	if !svc.trackMatcher.Match(expectedTrack, TrackForMatching{Artist: track.Artist, Title: track.Title}) {
		// check if we maybe switched up artist and title during guessing
		if !(sourceIsAutoManaged && svc.trackMatcher.Match(expectedTrack, TrackForMatching{Artist: track.Title, Title: track.Artist})) {
			actualText := fmt.Sprintf("artist=%s, title=%s", track.Artist, track.Title)
			expectedText := fmt.Sprintf("artist=%s, title=%s", expectedArtist, expectedTitle)
			return storage.Track{}, fmt.Errorf("track [%s] doesn't match expected [%s] with url=%s", actualText, expectedText, url)
		}
	}

	if sourceIsAutoManaged {
		track.Artist = expectedArtist
		track.Title = expectedTitle

		tracks, err = svc.tracks.ReplaceBySource(source.Id, []storage.Track{track})
		if err != nil {
			return storage.Track{}, fmt.Errorf("failed to fixup imported track info: %w", err)
		}
	}

	return tracks[0], nil
}
