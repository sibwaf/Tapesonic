package library

import "gorm.io/gorm"

type LibraryModule struct {
	coverStorage    *CoverStorage
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
	coverStorage := newCoverStorage(db)
	playlistStorage := newPlaylistStorage(db)

	service := newLibraryService(artistStorage, albumStorage, trackStorage, coverStorage, playlistStorage)

	return &LibraryModule{
		coverStorage:    coverStorage,
		artistStorage:   artistStorage,
		albumStorage:    albumStorage,
		trackStorage:    trackStorage,
		playlistStorage: playlistStorage,
		LibraryService:  service,
	}
}

func (module *LibraryModule) PrepareDatabase() error {
	if err := module.coverStorage.PrepareDatabase(); err != nil {
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
