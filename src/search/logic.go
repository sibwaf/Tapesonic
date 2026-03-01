package search

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"tapesonic/artists"
	"tapesonic/library"
	"tapesonic/model"
	"tapesonic/sources"
	"tapesonic/util"
	"tapesonic/ytdlp"

	"github.com/google/uuid"
)

type SearchService struct {
	library  *library.LibraryService
	artists  *artists.ArtistService
	sources  *sources.SourceService
	importer *AutoImportService
	matcher  *TrackMatcher
}

func newSearchService(
	library *library.LibraryService,
	artists *artists.ArtistService,
	sources *sources.SourceService,
	importer *AutoImportService,
	matcher *TrackMatcher,
) *SearchService {
	return &SearchService{
		library:  library,
		artists:  artists,
		sources:  sources,
		importer: importer,
		matcher:  matcher,
	}
}

func (svc *SearchService) FindTracksByQuery(userId uuid.UUID, query string) ([]model.LibraryTrack, error) {
	if query == "" {
		return []model.LibraryTrack{}, nil
	}

	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		tree, err := svc.importer.ImportTree(context.Background(), query, sources.SOURCE_MANAGEMENT_POLICY_MANUAL)
		if err != nil {
			return []model.LibraryTrack{}, err
		}

		ordering := map[string]int{}
		allIds := []string{}
		for i, child := range tree.CollectTracks() {
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
	match, err := svc.findMatchInLibrary(userId, track)
	if err != nil {
		return nil, err
	}
	if match != nil {
		return match, nil
	}

	trackForImport := TrackForImport{
		ArtistMusicBrainzId: track.ArtistMusicBrainzId,
		Artist:              track.Artist,
		Title:               track.Title,
	}

	for _, url := range track.Urls {
		importedTrack, err := svc.importer.ImportSingleTrack(context.Background(), url, trackForImport)
		if errors.Is(err, ytdlp.ErrNotAvailable) {
			slog.Debug(fmt.Sprintf("Failed to import track from %s: URL not available", url))
			continue
		} else if errors.Is(err, ErrMetadataMismatch) {
			continue
		} else if err != nil {
			return nil, err
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

	artists, err := svc.artists.FindAllMatches(track.Artist, track.ArtistMusicBrainzId)
	if err != nil {
		return nil, err
	}

	allowedArtistIds := []string{}
	for _, artist := range artists {
		allowedArtistIds = append(allowedArtistIds, artist.Id.String())
	}

	var match *model.LibraryTrack = nil
	for _, candidate := range candidates {
		if !slices.Contains(allowedArtistIds, candidate.ArtistId) || !util.MatchText(track.Title, candidate.Title) {
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
