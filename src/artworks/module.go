package artworks

import "gorm.io/gorm"

type ArtworksModule struct {
	storage        *ArtworkStorage
	ArtworkService *ArtworkService
}

func NewArtworksModule(db *gorm.DB, artworkDir string) *ArtworksModule {
	storage := newArtworkStorage(db)
	service := newArtworkService(storage, artworkDir)

	return &ArtworksModule{
		storage:        storage,
		ArtworkService: service,
	}
}

func (module *ArtworksModule) PrepareDatabase() error {
	return module.storage.PrepareDatabase()
}
