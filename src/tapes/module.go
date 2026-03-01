package tapes

import (
	"tapesonic/artists"

	"gorm.io/gorm"
)

type TapesModule struct {
	storage     *TapeStorage
	TapeService *TapeService
}

func NewTapesModule(
	db *gorm.DB,
	artists *artists.ArtistService,
) *TapesModule {
	storage := newTapeStorage(db)
	service := newTapeService(storage, artists)

	return &TapesModule{
		storage:     storage,
		TapeService: service,
	}
}

func (module *TapesModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
