package logic

import (
	"fmt"
	"strings"
	"tapesonic/storage"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type TapeService struct {
	tapes  *storage.TapeStorage
	tracks *storage.TrackStorage
}

func NewTapeService(
	tapes *storage.TapeStorage,
	tracks *storage.TrackStorage,
) *TapeService {
	return &TapeService{
		tapes:  tapes,
		tracks: tracks,
	}
}

func (s *TapeService) Create(tape storage.Tape) (storage.Tape, []storage.Track, error) {
	tape, err := s.tapes.Create(tape)

	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(tape.Id)
	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) Update(tape storage.Tape) (storage.Tape, []storage.Track, error) {
	tape, err := s.tapes.Update(tape)
	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(tape.Id)
	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) DeleteById(id uuid.UUID) error {
	return s.tapes.DeleteById(id)
}

func (s *TapeService) GetList() ([]storage.Tape, error) {
	return s.tapes.GetAllTapes()
}

func (s *TapeService) GetById(id uuid.UUID) (storage.Tape, []storage.Track, error) {
	tape, err := s.tapes.GetTape(id)
	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(id)
	if err != nil {
		return storage.Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) GuessTapeMetadata(trackIds []uuid.UUID) (storage.Tape, error) {
	tracks, err := s.tracks.GetTracksForTapeMetadataGuessing(trackIds)
	if err != nil {
		return storage.Tape{}, err
	}

	artists := util.NewCountingSet[string]()
	parentNames := util.NewCountingSet[string]()
	names := util.NewCountingSet[string]()
	releaseDates := util.NewCountingSet[time.Time]()
	thumbnailIds := util.NewCountingSet[uuid.UUID]()

	for _, track := range tracks {
		artists.Add(util.Coalesce(track.AlbumArtist, track.Artist))
		names.Add(util.Coalesce(track.AlbumTitle, track.SourceTitle))

		for _, parentName := range track.SourceParentTitles {
			parentNames.Add(parentName)
		}

		if track.ThumbnailId != nil {
			thumbnailIds.Add(*track.ThumbnailId)
		} else {
			thumbnailIds.Add(uuid.Nil)
		}

		if track.ReleaseDate != nil {
			releaseDates.Add(*track.ReleaseDate)
		} else {
			releaseDates.Add(time.Time{})
		}
	}

	threshold := float32(0.75)
	name := names.GetDominatingValue(threshold)
	parentName := parentNames.GetDominatingValue(threshold)
	artist := artists.GetDominatingValue(threshold)
	thumbnailId := thumbnailIds.GetDominatingValue(threshold)
	releaseDate := releaseDates.GetDominatingValue(threshold)

	result := storage.Tape{
		Name:        util.Coalesce(name, parentName),
		Artist:      artist,
		ReleasedAt:  util.TakeIf(&releaseDate, !releaseDate.IsZero()),
		ThumbnailId: util.TakeIf(&thumbnailId, thumbnailId.ID() != 0),
	}

	artistThenDash := fmt.Sprintf("%s - ", result.Artist)
	if strings.HasPrefix(strings.ToLower(result.Name), strings.ToLower(artistThenDash)) {
		result.Name = result.Name[len(artistThenDash):]
	}

	result.Name = strings.TrimSpace(result.Name)
	result.Artist = strings.TrimSpace(result.Artist)

	if artist != "" || !releaseDate.IsZero() {
		result.Type = storage.TAPE_TYPE_ALBUM
	} else {
		result.Type = storage.TAPE_TYPE_PLAYLIST
	}

	return result, nil
}
