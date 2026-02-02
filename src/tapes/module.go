package tapes

import (
	"gorm.io/gorm"
)

type TapesModule struct {
	storage     *TapeStorage
	TapeService *TapeService
}

func NewTapesModule(
	db *gorm.DB,
) *TapesModule {
	storage := newTapeStorage(db)
	service := newTapeService(storage)

	return &TapesModule{
		storage:     storage,
		TapeService: service,
	}
}

func (module *TapesModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
