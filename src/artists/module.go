package artists

import "gorm.io/gorm"

type ArtistsModule struct {
	storage       *artistStorage
	ArtistService *ArtistService
}

func NewArtistsModule(db *gorm.DB) *ArtistsModule {
	storage := newArtistStorage(db)

	return &ArtistsModule{
		storage:       storage,
		ArtistService: newArtistService(storage),
	}
}

func (module *ArtistsModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
