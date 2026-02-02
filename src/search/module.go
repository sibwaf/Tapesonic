package search

import (
	"tapesonic/library"
	"tapesonic/sources"

	"gorm.io/gorm"
)

type SearchModule struct {
	SearchService *SearchService
}

func NewSearchModule(
	db *gorm.DB,
	library *library.LibraryService,
	sources *sources.SourceService,
	importer *sources.ImportService,
) *SearchModule {
	trackMatcher := newTrackMatcher()
	autoImporter := newAutoImportService(importer, sources, trackMatcher)
	service := newSearchService(library, sources, importer, autoImporter, trackMatcher)

	return &SearchModule{
		SearchService: service,
	}
}
