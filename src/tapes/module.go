package tapes

import (
	"tapesonic/artists"
	"tapesonic/library"
	"tapesonic/sources"

	"gorm.io/gorm"
)

type TapesModule struct {
	storage     *TapeStorage
	TapeService *TapeService
}

func NewTapesModule(
	db *gorm.DB,
	artists *artists.ArtistService,
	sources *sources.SourceService,
	library *library.LibraryService,
) *TapesModule {
	storage := newTapeStorage(db)
	service := newTapeService(storage, artists, sources, library)

	return &TapesModule{
		storage:     storage,
		TapeService: service,
	}
}

func (module *TapesModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
