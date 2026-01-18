package tapes

import (
	"fmt"
	"strings"
	"tapesonic/model"
	"tapesonic/storage"
	"tapesonic/users"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type TapeService struct {
	tapes  *TapeStorage
	tracks *storage.TrackStorage
}

func newTapeService(
	tapes *TapeStorage,
	tracks *storage.TrackStorage,
) *TapeService {
	return &TapeService{
		tapes:  tapes,
		tracks: tracks,
	}
}

func (s *TapeService) Create(user users.User, tape Tape) (Tape, []storage.Track, error) {
	tape.Id = uuid.New()
	tape.CreatedById = user.Id
	tape.CreatedAt = util.NewTimestampWrapper(time.Now())
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	tape, err := s.tapes.Create(tape)

	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(tape.Id)
	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) Update(tape Tape) (Tape, []storage.Track, error) {
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	tape, err := s.tapes.Update(tape)
	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(tape.Id)
	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) DeleteById(id uuid.UUID) error {
	return s.tapes.DeleteById(id)
}

func (s *TapeService) GetList() ([]Tape, error) {
	return s.tapes.GetAllTapes()
}

func (s *TapeService) GetById(id uuid.UUID) (Tape, []storage.Track, error) {
	tape, err := s.tapes.GetTape(id)
	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	tracks, err := s.tracks.GetTracksByTape(id)
	if err != nil {
		return Tape{}, []storage.Track{}, err
	}

	return tape, tracks, nil
}

type TapeMetadata struct {
	Name        string
	Type        model.TapeType
	Artist      string
	ReleasedAt  *time.Time
	ThumbnailId *uuid.UUID
}

func (s *TapeService) GuessTapeMetadata(trackIds []uuid.UUID) (TapeMetadata, error) {
	tracks, err := s.tracks.GetTracksForTapeMetadataGuessing(trackIds)
	if err != nil {
		return TapeMetadata{}, err
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
			releaseDates.Add(track.ReleaseDate.Unwrap())
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

	result := TapeMetadata{
		Name:        util.Coalesce(name, parentName),
		Artist:      artist,
		ReleasedAt:  util.TakeIf(&releaseDate, !releaseDate.IsZero()),
		ThumbnailId: util.TakeIf(&thumbnailId, thumbnailId != uuid.Nil),
	}

	artistThenDash := fmt.Sprintf("%s - ", result.Artist)
	if strings.HasPrefix(strings.ToLower(result.Name), strings.ToLower(artistThenDash)) {
		result.Name = result.Name[len(artistThenDash):]
	}

	result.Name = strings.TrimSpace(result.Name)
	result.Artist = strings.TrimSpace(result.Artist)

	if artist != "" || !releaseDate.IsZero() {
		result.Type = model.TAPE_TYPE_ALBUM
	} else {
		result.Type = model.TAPE_TYPE_PLAYLIST
	}

	return result, nil
}
