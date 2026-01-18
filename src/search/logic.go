package search

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"tapesonic/library"
	"tapesonic/logic"
	"tapesonic/model"

	"github.com/google/uuid"
)

type SearchService struct {
	library  *library.LibraryService
	sources  *logic.SourceService
	tracks   *logic.TrackService
	importer *logic.AutoImportService
	matcher  *logic.TrackMatcher
}

func newSearchService(
	library *library.LibraryService,
	sources *logic.SourceService,
	tracks *logic.TrackService,
	importer *logic.AutoImportService,
	matcher *logic.TrackMatcher,
) *SearchService {
	return &SearchService{
		library:  library,
		sources:  sources,
		tracks:   tracks,
		importer: importer,
		matcher:  matcher,
	}
}

func (svc *SearchService) FindTracksByQuery(userId uuid.UUID, query string) ([]model.LibraryTrack, error) {
	if query == "" {
		return []model.LibraryTrack{}, nil
	}

	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		source, err := svc.sources.AddSource(context.Background(), query, model.SOURCE_MANAGEMENT_POLICY_MANUAL)
		if err != nil {
			return []model.LibraryTrack{}, err
		}

		sourceTracks, err := svc.tracks.GetAllTracksBySource(source.Id)
		if err != nil {
			return []model.LibraryTrack{}, err
		}

		ordering := map[string]int{}
		allIds := []string{}
		for i, child := range sourceTracks {
			idStr := child.Id.String()
			ordering[idStr] = i
			allIds = append(allIds, idStr)
		}

		// if something is stupid, but it works, then it is not stupid

		tracks, err := svc.library.GetTracksByIds(userId, allIds)
		if err != nil {
			return []model.LibraryTrack{}, err
		}

		slices.SortFunc(tracks, func(a model.LibraryTrack, b model.LibraryTrack) int {
			return ordering[a.Id] - ordering[b.Id]
		})

		return tracks, nil
	}

	// todo: limit
	return svc.library.SearchTracksByQuery(userId, query, 9999, 0)
}

func (svc *SearchService) FindTrack(userId uuid.UUID, track TrackForSearch) (*model.LibraryTrack, error) {
	expected := logic.TrackForMatching{
		Artist: track.Artist,
		Title:  track.Title,
	}

	match, err := svc.findMatchInLibrary(userId, track)
	if err != nil {
		return nil, err
	}
	if match != nil {
		return match, nil
	}

	for _, url := range track.Urls {
		importedTrack, err := svc.importer.ImportTrackFrom(context.Background(), url, expected.Artist, expected.Title)
		if err != nil {
			return nil, err
		}
		if importedTrack == nil {
			continue
		}

		track, err := svc.library.GetTrackById(userId, importedTrack.Id.String())
		return &track, err
	}

	return nil, nil
}

func (svc *SearchService) findMatchInLibrary(userId uuid.UUID, track TrackForSearch) (*model.LibraryTrack, error) {
	// todo: limit

	filter := library.TrackFilter{Artist: track.Artist, Title: track.Title, Album: track.Album}
	candidates, err := svc.library.SearchTracksByFields(userId, filter, 10, 0)
	if err != nil {
		return nil, err
	}

	expected := logic.TrackForMatching{
		Artist: track.Artist,
		Title:  track.Title,
	}

	var match *model.LibraryTrack = nil
	for _, candidate := range candidates {
		actual := logic.TrackForMatching{
			Artist: candidate.ArtistName,
			Title:  candidate.Title,
		}

		if !svc.matcher.Match(expected, actual) {
			continue
		}

		if match == nil {
			match = &candidate
		} else {
			// too many matches, abort for now
			slog.Debug(fmt.Sprintf("Found multiple matches in library while searching for track %+v, aborting", track))
			return nil, nil
		}
	}

	if match == nil && track.Album != "" {
		track.Album = ""
		return svc.findMatchInLibrary(userId, track)
	}

	return match, nil
}
