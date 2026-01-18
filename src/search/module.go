package search

import (
	"tapesonic/library"
	"tapesonic/logic"

	"gorm.io/gorm"
)

type SearchModule struct {
	SearchService *SearchService
}

func NewSearchModule(
	db *gorm.DB,
	library *library.LibraryService,
	sources *logic.SourceService,
	tracks *logic.TrackService,
	importer *logic.AutoImportService,
	trackMatcher *logic.TrackMatcher,
) *SearchModule {
	service := newSearchService(library, sources, tracks, importer, trackMatcher)

	return &SearchModule{
		SearchService: service,
	}
}
