package logic

import (
	"tapesonic/storage"

	"github.com/google/uuid"
)

type TrackService struct {
	storage *storage.TrackStorage
}

func NewTrackService(storage *storage.TrackStorage) *TrackService {
	return &TrackService{storage: storage}
}

func (s *TrackService) InitializeTracksFor(sourceId uuid.UUID, tracks []storage.Track) ([]storage.Track, error) {
	savedTracks, err := s.storage.GetDirectTracksBySource(sourceId)
	if err != nil {
		return []storage.Track{}, err
	}

	if len(savedTracks) > 0 {
		return savedTracks, nil
	}

	return s.ReplaceBySource(sourceId, tracks)
}

func (s *TrackService) ReplaceBySource(sourceId uuid.UUID, tracks []storage.Track) ([]storage.Track, error) {
	return s.storage.ReplaceTracksForSource(sourceId, tracks)
}

func (s *TrackService) GetDirectTracksBySource(sourceId uuid.UUID) ([]storage.Track, error) {
	return s.storage.GetDirectTracksBySource(sourceId)
}

func (s *TrackService) GetAllTracksBySource(sourceId uuid.UUID) ([]storage.Track, error) {
	return s.storage.GetAllTracksBySource(sourceId)
}
