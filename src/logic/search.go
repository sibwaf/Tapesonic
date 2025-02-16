package logic

import (
	"strings"
	"tapesonic/storage"
)

type SearchService struct {
	sources *storage.SourceStorage
	tracks  *storage.TrackStorage
}

func NewSearchService(
	sources *storage.SourceStorage,
	tracks *storage.TrackStorage,
) *SearchService {
	return &SearchService{
		sources: sources,
		tracks:  tracks,
	}
}

func (s *SearchService) SearchTracksForWebUi(query string) ([]storage.Track, error) {
	query = strings.TrimSpace(query)

	source, err := s.sources.FindByUrl(query)
	if err != nil {
		return []storage.Track{}, err
	}

	if source == nil {
		return []storage.Track{}, nil
	}

	return s.tracks.GetAllTracksBySource(source.Id)
}
