package tapes

import (
	"fmt"
	"slices"
	"strings"
	"tapesonic/artists"
	"tapesonic/library"
	"tapesonic/model"
	"tapesonic/sources"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type TapeService struct {
	tapes   *TapeStorage
	artists *artists.ArtistService
	sources *sources.SourceService
	library *library.LibraryService
}

func newTapeService(
	tapes *TapeStorage,
	artists *artists.ArtistService,
	sources *sources.SourceService,
	library *library.LibraryService,
) *TapeService {
	return &TapeService{
		tapes:   tapes,
		artists: artists,
		sources: sources,
		library: library,
	}
}

func (s *TapeService) Create(userId uuid.UUID, tape Tape) (SavedTape, []model.LibraryTrack, error) {
	tape.Id = uuid.New()
	tape.CreatedBy = userId
	tape.CreatedAt = util.NewTimestampWrapper(time.Now())
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	savedTape, err := s.tapes.Create(tape)

	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	trackIds := []string{}
	for _, track := range tape.Tracks {
		trackIds = append(trackIds, track.TrackId)
	}

	err = s.tapes.ReplaceTracksById(savedTape.Id, trackIds)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	tracks, err := s.fetchTracks(userId, savedTape.Id)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	return savedTape, tracks, nil
}

func (s *TapeService) Update(userId uuid.UUID, tapeId uuid.UUID, tape Tape) (SavedTape, []model.LibraryTrack, error) {
	tape.Id = tapeId
	tape.UpdatedAt = util.NewTimestampWrapper(time.Now())

	savedTape, err := s.tapes.Update(tape)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	trackIds := []string{}
	for _, track := range tape.Tracks {
		trackIds = append(trackIds, track.TrackId)
	}

	err = s.tapes.ReplaceTracksById(savedTape.Id, trackIds)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	tracks, err := s.fetchTracks(userId, tapeId)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	return savedTape, tracks, nil
}

func (s *TapeService) DeleteById(id uuid.UUID) error {
	return s.tapes.DeleteById(id)
}

func (s *TapeService) GetList() ([]SavedTape, error) {
	return s.tapes.GetAllTapes()
}

func (s *TapeService) GetById(userId uuid.UUID, tapeId uuid.UUID) (SavedTape, []model.LibraryTrack, error) {
	tape, err := s.tapes.GetById(tapeId)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	tracks, err := s.fetchTracks(userId, tapeId)
	if err != nil {
		return SavedTape{}, []model.LibraryTrack{}, err
	}

	return tape, tracks, nil
}

func (s *TapeService) fetchTracks(userId uuid.UUID, tapeId uuid.UUID) ([]model.LibraryTrack, error) {
	trackIds, err := s.tapes.GetTrackIdsById(tapeId)
	if err != nil {
		return []model.LibraryTrack{}, err
	}

	ordering := map[string]int{}
	for index, trackId := range trackIds {
		ordering[trackId] = index
	}

	tracks, err := s.library.GetTracksByIds(userId, trackIds)
	if err != nil {
		return []model.LibraryTrack{}, err
	}

	slices.SortFunc(tracks, func(a model.LibraryTrack, b model.LibraryTrack) int { return ordering[a.Id] - ordering[b.Id] })

	return tracks, nil
}

type TapeMetadata struct {
	Name       string
	Type       model.TapeType
	ArtistId   *uuid.UUID
	ArtistName string
	ReleasedAt *time.Time
	ArtworkId  *uuid.UUID
}

func (s *TapeService) GuessTapeMetadata(userId uuid.UUID, trackIds []string) (TapeMetadata, error) {
	libraryTracks, err := s.library.GetTracksByIds(userId, trackIds)
	if err != nil {
		return TapeMetadata{}, err
	}

	sourceTracks, err := s.sources.FindTracksForMetadataGuessingByIds(trackIds)
	if err != nil {
		return TapeMetadata{}, err
	}

	sourceInfoByTrackId := map[string]sources.SourceTrackForMetadataGuessing{}
	for _, sourceTrack := range sourceTracks {
		sourceInfoByTrackId[sourceTrack.Id.String()] = sourceTrack
	}

	parentNames := util.NewCountingSet[string]()
	names := util.NewCountingSet[string]()
	artistIds := util.NewCountingSet[uuid.UUID]()
	releaseDates := util.NewCountingSet[time.Time]()
	artworkIds := util.NewCountingSet[uuid.UUID]()

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

	for _, track := range libraryTracks {
		sourceTrack := sourceInfoByTrackId[track.Id]

		names.Add(util.Coalesce(sourceTrack.AlbumTitle, sourceTrack.SourceTitle, track.AlbumName))

		for _, parentName := range sourceTrack.SourceParentTitles {
			parentNames.Add(parentName)
		}
		if len(sourceTrack.SourceParentTitles) == 0 {
			parentNames.Add("")
		}

		if sourceTrack.AlbumArtist != "" {
			artistId, err := getArtistId(strings.TrimSpace(sourceTrack.AlbumArtist))
			if err != nil {
				return TapeMetadata{}, err
			}

			artistIds.Add(artistId)
		} else if track.AlbumArtistId != nil {
			artistIds.Add(*track.AlbumArtistId)
		} else if track.ArtistId != nil {
			artistIds.Add(*track.ArtistId)
		} else {
			artistIds.Add(uuid.Nil)
		}

		if sourceTrack.ArtworkId != nil {
			artworkIds.Add(*sourceTrack.ArtworkId)
		} else if track.ArtworkId != nil {
			artworkIds.Add(*track.ArtworkId)
		} else {
			artworkIds.Add(uuid.Nil)
		}

		if sourceTrack.ReleaseDate != nil {
			releaseDates.Add(sourceTrack.ReleaseDate.Unwrap())
		} else if track.AlbumReleasedAt != nil {
			releaseDates.Add(track.AlbumReleasedAt.Unwrap())
		} else {
			releaseDates.Add(time.Time{})
		}
	}

	threshold := float32(0.75)
	name := names.GetDominatingValue(threshold)
	parentName := parentNames.GetDominatingValue(threshold)
	artistId := artistIds.GetDominatingValue(threshold)
	artworkId := artworkIds.GetDominatingValue(threshold)
	releaseDate := releaseDates.GetDominatingValue(threshold)

	result := TapeMetadata{
		Name:       util.Coalesce(name, parentName),
		ReleasedAt: util.TakeIf(&releaseDate, !releaseDate.IsZero()),
		ArtworkId:  util.TakeIf(&artworkId, artworkId != uuid.Nil),
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
