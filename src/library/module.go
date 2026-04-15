package library

import "gorm.io/gorm"

type LibraryModule struct {
	artworkStorage  *ArtworkStorage
	artistStorage   *ArtistStorage
	albumStorage    *AlbumStorage
	trackStorage    *TrackStorage
	playlistStorage *PlaylistStorage

	LibraryService *LibraryService
}

func NewLibraryModule(db *gorm.DB) *LibraryModule {
	artistStorage := newArtistStorage(db)
	albumStorage := newAlbumStorage(db)
	trackStorage := newTrackStorage(db)
	artworkStorage := newArtworkStorage(db)
	playlistStorage := newPlaylistStorage(db)

	service := newLibraryService(artistStorage, albumStorage, trackStorage, artworkStorage, playlistStorage)

	return &LibraryModule{
		artworkStorage:  artworkStorage,
		artistStorage:   artistStorage,
		albumStorage:    albumStorage,
		trackStorage:    trackStorage,
		playlistStorage: playlistStorage,
		LibraryService:  service,
	}
}

func (module *LibraryModule) PrepareDatabase() error {
	if err := module.artworkStorage.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.artistStorage.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.albumStorage.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.trackStorage.PrepareDatabase(); err != nil {
		return err
	}
	if err := module.playlistStorage.PrepareDatabase(); err != nil {
		return err
	}
	return nil
}
