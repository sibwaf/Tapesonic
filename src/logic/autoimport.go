package logic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/model"
	"tapesonic/storage"
	"tapesonic/ytdlp"
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

func (svc *AutoImportService) ImportTrackFrom(ctx context.Context, url string, expectedArtist string, expectedTitle string) (*storage.Track, error) {
	source, err := svc.sources.FindByUrl(url)
	if err != nil {
		return nil, err
	}

	tracks := []storage.Track{}
	if source != nil {
		tracks, err = svc.tracks.GetDirectTracksBySource(source.Id)
		if err != nil {
			return nil, err
		}
	}

	expectedTrack := TrackForMatching{Artist: expectedArtist, Title: expectedTitle}

	if source == nil || len(tracks) == 0 {
		importedSource, err := svc.sources.AddSource(ctx, url, model.SOURCE_MANAGEMENT_POLICY_AUTO)
		if errors.Is(err, ytdlp.ErrNotAvailable) {
			slog.Warn(fmt.Sprintf("Auto-import for %+v failed: %s is not available", expectedTrack, url))
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("failed to import source: %w", err)
		}

		tracks, err = svc.tracks.GetDirectTracksBySource(importedSource.Id)
		if err != nil {
			return nil, err
		}

		source = &importedSource
	}

	if len(tracks) == 0 {
		slog.Warn(fmt.Sprintf("Auto-import for %+v failed: no tracks were imported from %s", expectedTrack, url))
		return nil, nil
	} else if len(tracks) > 1 {
		slog.Warn(fmt.Sprintf("Auto-import for %+v failed: multiple tracks were imported from %s", expectedTrack, url))
		return nil, nil
	}

	sourceIsAutoManaged := source.ManagementPolicy == model.SOURCE_MANAGEMENT_POLICY_AUTO
	track := tracks[0]

	if !svc.trackMatcher.Match(expectedTrack, TrackForMatching{Artist: track.Artist, Title: track.Title}) {
		// check if we maybe switched up artist and title during guessing
		if !(sourceIsAutoManaged && svc.trackMatcher.Match(expectedTrack, TrackForMatching{Artist: track.Title, Title: track.Artist})) {
			actualTrack := TrackForMatching{Artist: track.Artist, Title: track.Title}
			slog.Debug(fmt.Sprintf("Auto-imported track %+v from %s doesn't match expected %+v", actualTrack, url, expectedTrack))
			return nil, nil
		}
	}

	if sourceIsAutoManaged {
		track.Artist = expectedArtist
		track.Title = expectedTitle

		tracks, err = svc.tracks.ReplaceBySource(source.Id, []storage.Track{track})
		if err != nil {
			return nil, err
		}
	}

	return &tracks[0], nil
}
