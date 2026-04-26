package scrobbling

import (
	"tapesonic/lastfm"
	"tapesonic/library"
	"tapesonic/listenbrainz"
	"tapesonic/remotes"

	"gorm.io/gorm"
)

type ScrobblingModule struct {
	listenStatStorage *ListenStatStorage
	ScrobbleService   *ScrobbleService
}

func NewScrobblingModule(
	db *gorm.DB,
	listenbrainz *listenbrainz.ListenBrainzService,
	lastfm *lastfm.LastFmService,
	library *library.LibraryService,
	remotes *remotes.RemoteService,
) *ScrobblingModule {
	statStorage := newListenStatStorage(db)

	return &ScrobblingModule{
		listenStatStorage: statStorage,
		ScrobbleService:   NewScrobbleService(listenbrainz, lastfm, library, remotes, statStorage),
	}
}
