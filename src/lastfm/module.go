package lastfm

import "gorm.io/gorm"

type LastFmModule struct {
	LastFmService               *LastFmService
	LastFmRecommendationService *LastFmRecommendationService
}

func NewLastFmModule(
	db *gorm.DB,
	apiKey string,
	apiSecret string,
) (*LastFmModule, error) {
	sessions, err := newLastFmSessionStorage(db)
	if err != nil {
		return nil, err
	}

	var client *LastFmClient = nil
	if apiKey != "" && apiSecret != "" {
		client = newLastFmClient(apiKey, apiSecret)
	}

	service := newLastFmService(client, sessions)
	recommendationService := newLastFmRecommendationService(sessions, client)

	return &LastFmModule{
		LastFmService:               service,
		LastFmRecommendationService: recommendationService,
	}, nil
}
