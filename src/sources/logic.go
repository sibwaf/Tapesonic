package sources

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/google/uuid"
)

type SourceService struct {
	sources *SourceStorage
	tracks  *TrackStorage
	files   *FileStorage

	baseDir string
}

func newSourceService(
	sources *SourceStorage,
	tracks *TrackStorage,
	files *FileStorage,
	baseDir string,
) *SourceService {
	return &SourceService{
		sources: sources,
		tracks:  tracks,
		files:   files,
		baseDir: baseDir,
	}
}

func (svc *SourceService) GetListForApi(managementPolicies []SourceManagementPolicy) ([]SourceForApi, error) {
	sources, err := svc.sources.GetListForApi(managementPolicies)
	if err != nil {
		return []SourceForApi{}, err
	}

	sourceIds := []uuid.UUID{}
	for _, source := range sources {
		sourceIds = append(sourceIds, source.Id)
	}

	files, err := svc.files.FindBySourceIds(sourceIds)
	if err != nil {
		return []SourceForApi{}, err
	}

	fileLookup := map[uuid.UUID]SourceFile{}
	for _, file := range files {
		fileLookup[file.SourceId] = file
	}

	result := []SourceForApi{}
	for _, source := range sources {
		dto := SourceForApi{Source: source}
		if file, ok := fileLookup[source.Id]; ok {
			dto.File = &file
		}

		result = append(result, dto)
	}

	return result, nil
}

func (svc *SourceService) GetByIdForApi(id uuid.UUID) (SourceForApi, error) {
	source, err := svc.sources.GetById(id)
	if err != nil {
		return SourceForApi{}, err
	}

	file, err := svc.files.FindBySourceId(id)
	if err != nil {
		return SourceForApi{}, err
	}

	return SourceForApi{
		Source: source,
		File:   file,
	}, nil
}

func (svc *SourceService) FindByUrl(url string) (*Source, error) {
	return svc.sources.FindByUrl(url)
}

func (svc *SourceService) GetHierarchy(id uuid.UUID) ([]SourceForHierarchy, error) {
	return svc.sources.GetHierarchy(id)
}

func (svc *SourceService) ReplaceTracks(sourceId uuid.UUID, tracks []SourceTrack, managementPolicy SourceManagementPolicy) ([]SavedSourceTrack, error) {
	currentManagementPolicy, err := svc.sources.GetManagementPolicyById(sourceId)
	if err != nil {
		return []SavedSourceTrack{}, err
	}

	if currentManagementPolicy == SOURCE_MANAGEMENT_POLICY_MANUAL && managementPolicy != SOURCE_MANAGEMENT_POLICY_MANUAL {
		return svc.tracks.GetDirectTracksBySource(sourceId)
	}

	if managementPolicy == SOURCE_MANAGEMENT_POLICY_MANUAL && currentManagementPolicy != managementPolicy {
		if err := svc.sources.SetManagementPolicyById(sourceId, managementPolicy); err != nil {
			return []SavedSourceTrack{}, fmt.Errorf("failed to update source management policy: %w", err)
		}
	}

	for i := range tracks {
		if tracks[i].Id == uuid.Nil {
			tracks[i].Id = uuid.New()
		}
	}

	return svc.tracks.ReplaceTracksForSource(sourceId, tracks)
}

func (svc *SourceService) GetDirectTracks(sourceId uuid.UUID) ([]SavedSourceTrack, error) {
	return svc.tracks.GetDirectTracksBySource(sourceId)
}

func (svc *SourceService) GetAllTracks(sourceId uuid.UUID) ([]SavedSourceTrack, error) {
	return svc.tracks.GetAllTracksBySource(sourceId)
}

func (svc *SourceService) DeleteFile(sourceId uuid.UUID) error {
	slog.Debug(fmt.Sprintf("Trying to delete media for source id=%s", sourceId))

	file, err := svc.files.FindBySourceId(sourceId)
	if err != nil {
		return err
	}

	if file == nil {
		slog.Debug(fmt.Sprintf("No file metadata found for source id=%s, nothing to delete", sourceId))
		return nil
	}

	mediaPath := path.Join(svc.baseDir, file.MediaPath)
	slog.Debug(fmt.Sprintf("Deleting file id=%s (%s) for source id=%s", file.Id, mediaPath, sourceId))

	err = os.Remove(file.MediaPath)
	if errors.Is(err, os.ErrNotExist) {
		slog.Debug(fmt.Sprintf("File id=%s (%s) for source id=%s doesn't exist in FS, deleting metadata", file.Id, mediaPath, sourceId))
	} else if err != nil {
		return err
	}

	err = svc.files.DeleteById(file.Id)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Deleted file id=%s (%s) for source id=%s", file.Id, mediaPath, sourceId))
	return nil
}
