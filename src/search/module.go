package search

import (
	"tapesonic/artists"
	"tapesonic/library"
	"tapesonic/sources"
)

type SearchModule struct {
	SearchService *SearchService
}

func NewSearchModule(
	artists *artists.ArtistService,
	library *library.LibraryService,
	sources *sources.SourceService,
	importer *sources.ImportService,
) *SearchModule {
	trackNormalizer := NewTrackNormalizer()
	trackMatcher := newTrackMatcher()
	autoImporter := newAutoImportService(artists, importer, sources, trackNormalizer, trackMatcher)
	service := newSearchService(library, artists, sources, autoImporter, trackMatcher)

	return &SearchModule{
		SearchService: service,
	}
}
