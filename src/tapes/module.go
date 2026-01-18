package tapes

import (
	"tapesonic/storage"

	"gorm.io/gorm"
)

type TapesModule struct {
	storage     *TapeStorage
	TapeService *TapeService
}

func NewTapesModule(
	db *gorm.DB,
	tracks *storage.TrackStorage,
) *TapesModule {
	storage := newTapeStorage(db)
	service := newTapeService(storage, tracks)

	return &TapesModule{
		storage:     storage,
		TapeService: service,
	}
}

func (module *TapesModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
