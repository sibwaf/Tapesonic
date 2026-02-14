package search

import (
	"context"
	"fmt"
	"log/slog"
	"tapesonic/sources"
)

type AutoImportService struct {
	importer   *sources.ImportService
	sources    *sources.SourceService
	normalizer *TrackNormalizer
	matcher    *TrackMatcher
}

func newAutoImportService(
	importer *sources.ImportService,
	sources *sources.SourceService,
	normalizer *TrackNormalizer,
	matcher *TrackMatcher,
) *AutoImportService {
	return &AutoImportService{
		importer:   importer,
		sources:    sources,
		normalizer: normalizer,
		matcher:    matcher,
	}
}

func (svc *AutoImportService) ImportTree(ctx context.Context, url string, managementPolicy sources.SourceManagementPolicy) (sources.SourceTree, error) {
	analyzedTree, err := svc.importer.Analyze(ctx, url)
	if err != nil {
		return sources.SourceTree{}, err
	}

	analyzedTree, err = svc.normalizer.NormalizeTree(analyzedTree)
	if err != nil {
		return sources.SourceTree{}, err
	}

	return svc.importer.ImportTree(analyzedTree, managementPolicy)
}

func (svc *AutoImportService) ImportSingleTrack(ctx context.Context, url string, expectedArtist string, expectedTitle string) (sources.SourceTrack, error) {
	analyzedTree, err := svc.importer.Analyze(ctx, url)
	if err != nil {
		return sources.SourceTrack{}, err
	}

	tracks, err := svc.normalizer.Normalize(analyzedTree.Tracks)
	if err != nil {
		return sources.SourceTrack{}, err
	}
	if len(tracks) != 1 {
		slog.Debug(fmt.Sprintf("%s must contain a single track, got count=%d", url, len(tracks)))
		return sources.SourceTrack{}, ErrMetadataMismatch
	}

	track := tracks[0]
	expectedTrack := TrackForMatching{Artist: expectedArtist, Title: expectedTitle}

	if !svc.matcher.Match(expectedTrack, TrackForMatching{Artist: track.Artist, Title: track.Title}) {
		// check if we maybe switched up artist and title during guessing
		if !svc.matcher.Match(expectedTrack, TrackForMatching{Artist: track.Title, Title: track.Artist}) {
			actualTrack := TrackForMatching{Artist: track.Artist, Title: track.Title}
			slog.Debug(fmt.Sprintf("Track %+v from %s doesn't match expected %+v", actualTrack, url, expectedTrack))
			return sources.SourceTrack{}, ErrMetadataMismatch
		}
	}

	track.Artist = expectedArtist
	track.Title = expectedTitle
	analyzedTree.Tracks = []sources.TrackProperties{track}

	importedTree, err := svc.importer.ImportTree(analyzedTree, sources.SOURCE_MANAGEMENT_POLICY_AUTO)
	if err != nil {
		return sources.SourceTrack{}, err
	}

	if len(importedTree.Tracks) != 1 {
		slog.Debug(fmt.Sprintf("%s must contain a single track, got count=%d after importing", url, len(importedTree.Tracks)))
		return sources.SourceTrack{}, ErrMetadataMismatch
	}

	importedTrack := importedTree.Tracks[0]
	if svc.matcher.Match(expectedTrack, TrackForMatching{Artist: importedTrack.Artist, Title: importedTrack.Title}) {
		return importedTrack, nil
	} else {
		actualTrack := TrackForMatching{Artist: track.Artist, Title: track.Title}
		slog.Debug(fmt.Sprintf("Track %+v from %s doesn't match expected %+v after importing", actualTrack, url, expectedTrack))
		return sources.SourceTrack{}, ErrMetadataMismatch
	}
}
