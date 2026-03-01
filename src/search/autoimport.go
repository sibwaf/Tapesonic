package search

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"tapesonic/artists"
	"tapesonic/sources"
	"tapesonic/util"

	"github.com/google/uuid"
)

type AutoImportService struct {
	artists    *artists.ArtistService
	importer   *sources.ImportService
	sources    *sources.SourceService
	normalizer *TrackNormalizer
	matcher    *TrackMatcher
}

func newAutoImportService(
	artists *artists.ArtistService,
	importer *sources.ImportService,
	sources *sources.SourceService,
	normalizer *TrackNormalizer,
	matcher *TrackMatcher,
) *AutoImportService {
	return &AutoImportService{
		artists:    artists,
		importer:   importer,
		sources:    sources,
		normalizer: normalizer,
		matcher:    matcher,
	}
}

func (svc *AutoImportService) ImportTree(ctx context.Context, url string, managementPolicy sources.SourceManagementPolicy) (sources.SourceTree, error) {
	tree, err := svc.importer.Analyze(ctx, url)
	if err != nil {
		return sources.SourceTree{}, err
	}

	tree, err = svc.normalizer.NormalizeTree(tree)
	if err != nil {
		return sources.SourceTree{}, err
	}

	artistMapping := map[string]uuid.UUID{}
	getArtistId := func(name string) (uuid.UUID, error) {
		if artistId, ok := artistMapping[name]; ok {
			return artistId, nil
		}

		artist, err := svc.artists.FindMatchOrCreate(name, "")
		if err != nil {
			return uuid.UUID{}, err
		}

		artistMapping[name] = artist.Id
		return artist.Id, nil
	}

	var postprocessTree func(sources.AnalyzedSourceTree) (sources.AnalyzedSourceTree, error)
	postprocessTree = func(node sources.AnalyzedSourceTree) (sources.AnalyzedSourceTree, error) {
		for i, child := range node.Children {
			child, err := postprocessTree(child)
			if err != nil {
				return sources.AnalyzedSourceTree{}, err
			}
			node.Children[i] = child
		}

		for i, track := range node.Tracks {
			if track.ArtistId == nil && track.Artist != "" {
				artistId, err := getArtistId(track.Artist)
				if err != nil {
					return sources.AnalyzedSourceTree{}, err
				}
				track.ArtistId = &artistId
			}
			node.Tracks[i] = track
		}

		return node, nil
	}

	tree, err = postprocessTree(tree)
	if err != nil {
		return sources.SourceTree{}, err
	}

	return svc.importer.ImportTree(tree, managementPolicy)
}

func (svc *AutoImportService) ImportSingleTrack(ctx context.Context, url string, expectedTrack TrackForImport) (sources.SavedSourceTrack, error) {
	analyzedTree, err := svc.importer.Analyze(ctx, url)
	if err != nil {
		return sources.SavedSourceTrack{}, err
	}

	tracks, err := svc.normalizer.Normalize(analyzedTree.Tracks)
	if err != nil {
		return sources.SavedSourceTrack{}, err
	}
	if len(tracks) != 1 {
		slog.Debug(fmt.Sprintf("%s must contain a single track, got count=%d", url, len(tracks)))
		return sources.SavedSourceTrack{}, ErrMetadataMismatch
	}

	track := tracks[0]

	track, err = svc.checkTrackMatchesExpected(expectedTrack, track)
	if errors.Is(err, ErrMetadataMismatch) {
		actualTrack := TrackForMatching{Artist: track.Artist, Title: track.Title}
		slog.Debug(fmt.Sprintf("Track %+v from %s doesn't match expected %+v", actualTrack, url, expectedTrack))
		return sources.SavedSourceTrack{}, ErrMetadataMismatch
	} else if err != nil {
		return sources.SavedSourceTrack{}, err
	}

	artist, err := svc.artists.FindMatchOrCreate(expectedTrack.Artist, expectedTrack.ArtistMusicBrainzId)
	if err != nil {
		return sources.SavedSourceTrack{}, err
	}

	track.ArtistId = &artist.Id
	track.Title = expectedTrack.Title
	analyzedTree.Tracks = []sources.TrackProperties{track}

	importedTree, err := svc.importer.ImportTree(analyzedTree, sources.SOURCE_MANAGEMENT_POLICY_AUTO)
	if err != nil {
		return sources.SavedSourceTrack{}, err
	}

	if len(importedTree.Tracks) != 1 {
		slog.Debug(fmt.Sprintf("%s must contain a single track, got count=%d after importing", url, len(importedTree.Tracks)))
		return sources.SavedSourceTrack{}, ErrMetadataMismatch
	}

	importedTrack := importedTree.Tracks[0]

	if importedTrack.ArtistId == nil || *importedTrack.ArtistId != artist.Id || !util.MatchText(expectedTrack.Title, importedTrack.Title) {
		actualTrack := TrackForMatching{Artist: track.Artist, Title: track.Title}
		slog.Debug(fmt.Sprintf("Track %+v from %s doesn't match expected %+v after importing", actualTrack, url, expectedTrack))
		return sources.SavedSourceTrack{}, ErrMetadataMismatch
	}

	return importedTrack, nil
}

func (svc *AutoImportService) checkTrackMatchesExpected(expected TrackForImport, actual sources.TrackProperties) (sources.TrackProperties, error) {
	possibleArtists, err := svc.artists.FindAllMatches(expected.Artist, expected.ArtistMusicBrainzId)
	if err != nil {
		return sources.TrackProperties{}, err
	}

	allowedArtistNames := []string{expected.Artist}
	for _, possibleArtist := range possibleArtists {
		allowedArtistNames = append(allowedArtistNames, possibleArtist.Name)
		allowedArtistNames = append(allowedArtistNames, possibleArtist.Aliases...)
	}

	for _, allowedArtistName := range allowedArtistNames {
		expectedTrackForMatching := TrackForMatching{Artist: allowedArtistName, Title: expected.Title}

		if svc.matcher.Match(
			expectedTrackForMatching,
			TrackForMatching{Artist: actual.Artist, Title: actual.Title},
		) {
			return actual, nil
		}

		if svc.matcher.Match(
			expectedTrackForMatching,
			TrackForMatching{Artist: actual.Title, Title: actual.Artist},
		) {
			actual.Artist, actual.Title = actual.Title, actual.Artist
			return actual, nil
		}
	}

	return actual, ErrMetadataMismatch
}
