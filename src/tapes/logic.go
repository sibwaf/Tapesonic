package tapes

import (
	"fmt"
	"strings"
	"tapesonic/artists"
	"tapesonic/model"
	"tapesonic/sources"
	"tapesonic/users"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type TapeService struct {
	tapes   *TapeStorage
	artists *artists.ArtistService
}

func newTapeService(
	tapes *TapeStorage,
	artists *artists.ArtistService,
) *TapeService {
	return &TapeService{
		tapes:   tapes,
		artists: artists,
	}
}

func (s *TapeService) Create(user users.User, tape Tape) (SavedTape, []sources.SavedSourceTrack, error) {
	tape.Id = uuid.New()
	tape.CreatedById = user.Id
	tape.CreatedAt = util.NewTimestampWrapper(time.Now())
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	savedTape, err := s.tapes.Create(tape)

	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	trackIds := []uuid.UUID{}
	for _, track := range tape.Tracks {
		trackIds = append(trackIds, track.TrackId)
	}

	err = s.tapes.ReplaceTracksById(savedTape.Id, trackIds)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	tracks, err := s.tapes.GetTracksById(savedTape.Id)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	return savedTape, tracks, nil
}

func (s *TapeService) Update(id uuid.UUID, tape Tape) (SavedTape, []sources.SavedSourceTrack, error) {
	tape.Id = id
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	savedTape, err := s.tapes.Update(tape)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	trackIds := []uuid.UUID{}
	for _, track := range tape.Tracks {
		trackIds = append(trackIds, track.TrackId)
	}

	err = s.tapes.ReplaceTracksById(savedTape.Id, trackIds)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	tracks, err := s.tapes.GetTracksById(savedTape.Id)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	return savedTape, tracks, nil
}

func (s *TapeService) DeleteById(id uuid.UUID) error {
	return s.tapes.DeleteById(id)
}

func (s *TapeService) GetList() ([]SavedTape, error) {
	return s.tapes.GetAllTapes()
}

func (s *TapeService) GetById(id uuid.UUID) (SavedTape, []sources.SavedSourceTrack, error) {
	tape, err := s.tapes.GetById(id)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	tracks, err := s.tapes.GetTracksById(id)
	if err != nil {
		return SavedTape{}, []sources.SavedSourceTrack{}, err
	}

	return tape, tracks, nil
}

type TapeMetadata struct {
	Name        string
	Type        model.TapeType
	ArtistId    *uuid.UUID
	ArtistName  string
	ReleasedAt  *time.Time
	ThumbnailId *uuid.UUID
}

func (s *TapeService) GuessTapeMetadata(trackIds []uuid.UUID) (TapeMetadata, error) {
	tracks, err := s.tapes.GetTracksForMetadataGuessing(trackIds)
	if err != nil {
		return TapeMetadata{}, err
	}

	parentNames := util.NewCountingSet[string]()
	names := util.NewCountingSet[string]()
	artistIds := util.NewCountingSet[uuid.UUID]()
	releaseDates := util.NewCountingSet[time.Time]()
	thumbnailIds := util.NewCountingSet[uuid.UUID]()

	artistMapping := map[string]uuid.UUID{}
	getArtistId := func(name string) (uuid.UUID, error) {
		if artistId, ok := artistMapping[name]; ok {
			return artistId, nil
		}

		artist, err := s.artists.FindMatchOrCreate(name, "") // todo: wtf
		if err != nil {
			return uuid.UUID{}, err
		}

		artistMapping[name] = artist.Id
		return artist.Id, nil
	}

	for _, track := range tracks {
		names.Add(util.Coalesce(track.AlbumTitle, track.SourceTitle))

		for _, parentName := range track.SourceParentTitles {
			parentNames.Add(parentName)
		}

		if track.AlbumArtist != "" {
			artistId, err := getArtistId(strings.TrimSpace(track.AlbumArtist))
			if err != nil {
				return TapeMetadata{}, err
			}

			artistIds.Add(artistId)
		} else if track.ArtistId != nil {
			artistIds.Add(*track.ArtistId)
		} else {
			artistIds.Add(uuid.Nil)
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
	artistId := artistIds.GetDominatingValue(threshold)
	thumbnailId := thumbnailIds.GetDominatingValue(threshold)
	releaseDate := releaseDates.GetDominatingValue(threshold)

	result := TapeMetadata{
		Name:        util.Coalesce(name, parentName),
		ReleasedAt:  util.TakeIf(&releaseDate, !releaseDate.IsZero()),
		ThumbnailId: util.TakeIf(&thumbnailId, thumbnailId != uuid.Nil),
	}

	if artistId != uuid.Nil {
		artist, err := s.artists.GetById(artistId)
		if err != nil {
			return TapeMetadata{}, err
		}

		result.ArtistId = &artist.Id
		result.ArtistName = artist.Name

		artistThenDash := fmt.Sprintf("%s - ", artist.Name)
		if strings.HasPrefix(strings.ToLower(result.Name), strings.ToLower(artistThenDash)) {
			result.Name = result.Name[len(artistThenDash):]
		}
	}

	result.Name = strings.TrimSpace(result.Name)

	if result.ArtistId != nil || result.ReleasedAt != nil {
		result.Type = model.TAPE_TYPE_ALBUM
	} else {
		result.Type = model.TAPE_TYPE_PLAYLIST
	}

	return result, nil
}
