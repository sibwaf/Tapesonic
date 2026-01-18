package listenbrainz

import "gorm.io/gorm"

type ListenBrainzModule struct {
	sessions                          *ListenBrainzSessionStorage
	ListenBrainzService               *ListenBrainzService
	ListenBrainzRecommendationService *ListenBrainzRecommendationService
}

func NewListenBrainzModule(db *gorm.DB) *ListenBrainzModule {
	sessions := newListenBrainzSessionStorage(db)
	client := newListenBrainzClient()
	service := newListenBrainzService(sessions, client)
	recommendationService := newListenBrainzRecommendationService(sessions, client)

	return &ListenBrainzModule{
		sessions:                          sessions,
		ListenBrainzService:               service,
		ListenBrainzRecommendationService: recommendationService,
	}
}

func (module *ListenBrainzModule) PrepareDatabase() error {
	return module.sessions.PrepareDatabase()
}
