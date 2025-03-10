package logic

import (
	"context"
	"fmt"
	"tapesonic/model"
	"tapesonic/storage"
	"tapesonic/util"
	"tapesonic/ytdlp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type SourceService struct {
	storage *storage.SourceStorage

	ytdlp      *YtdlpService
	files      *SourceFileService
	tracks     *TrackService
	thumbnails *ThumbnailService

	normalizer *TrackNormalizer
}

func NewSourceService(
	storage *storage.SourceStorage,
	ytdlp *YtdlpService,
	files *SourceFileService,
	tracks *TrackService,
	thumbnails *ThumbnailService,
	normalizer *TrackNormalizer,
) *SourceService {
	return &SourceService{
		storage:    storage,
		ytdlp:      ytdlp,
		files:      files,
		tracks:     tracks,
		thumbnails: thumbnails,
		normalizer: normalizer,
	}
}

func (s *SourceService) AddSource(ctx context.Context, url string, managementPolicy model.SourceManagementPolicy) (storage.Source, error) {
	result, err := s.addSourceRecursive(ctx, url, managementPolicy, uuid.Nil)
	return result.Source, err
}

type SourceAndMetadata struct {
	Source   storage.Source
	Metadata ytdlp.YtdlpFile
}

func (s *SourceService) addSourceRecursive(ctx context.Context, url string, managementPolicy model.SourceManagementPolicy, parentId uuid.UUID) (SourceAndMetadata, error) {
	metadata, err := s.ytdlp.GetMetadata(ctx, url)
	if err != nil {
		return SourceAndMetadata{}, err
	}

	var thumbnail *storage.Thumbnail = nil
	if metadata.Thumbnail != "" {
		savedThumbnail, err := s.thumbnails.CreateFromUrl(metadata.Thumbnail)
		if err != nil {
			return SourceAndMetadata{}, err
		}

		thumbnail = &savedThumbnail
	}

	// this really needs SELECT FOR UPDATE in a transaction, but oh well
	existingSource, err := s.storage.FindByUrl(metadata.WebpageUrl)
	if err != nil {
		return SourceAndMetadata{}, err
	}

	source := storage.Source{
		ExtractorKey: metadata.ExtractorKey,
		ExtractedId:  metadata.Id,
		Url:          metadata.WebpageUrl,

		Title:      metadata.Title,
		Uploader:   metadata.Uploader,
		UploaderId: metadata.UploaderId,

		AlbumArtist: metadata.AlbumArtist,
		AlbumTitle:  metadata.Album,
		AlbumIndex:  metadata.TrackNumber,
		TrackArtist: metadata.Artist,
		TrackTitle:  metadata.Track,
		DurationMs:  int64(metadata.Duration * 1000),

		UploadedAt:  time.Unix(int64(metadata.Timestamp), 0),
		ReleaseDate: parseDateOrNull(metadata.ReleaseDate),

		Thumbnail: thumbnail,
	}

	// never override MANUAL management policy
	if existingSource == nil || existingSource.ManagementPolicy != model.SOURCE_MANAGEMENT_POLICY_MANUAL {
		source.ManagementPolicy = managementPolicy
	} else {
		source.ManagementPolicy = existingSource.ManagementPolicy
	}

	source, err = s.storage.Upsert(source)
	if err != nil {
		return SourceAndMetadata{}, err
	}

	tracks := []TrackProperties{}

	if len(metadata.Entries) > 0 {
		wg, nestedCtx := errgroup.WithContext(ctx)

		children := make([]SourceAndMetadata, len(metadata.Entries))
		for i := range metadata.Entries {
			index := i
			wg.Go(func() error {
				entryUrl := util.Coalesce(metadata.Entries[index].WebpageUrl, metadata.Entries[index].Url)

				child, err := s.addSourceRecursive(nestedCtx, entryUrl, managementPolicy, source.Id)
				if err != nil {
					return err
				}

				children[index] = child
				return nil
			})
		}

		if err := wg.Wait(); err != nil {
			return SourceAndMetadata{Source: source, Metadata: metadata}, fmt.Errorf("failed to add nested entry: %w", err)
		}

		childIds := make([]uuid.UUID, len(children))
		for i, child := range children {
			childIds[i] = child.Source.Id
		}

		if err := s.storage.UpdateHierarchy(source.Id, childIds); err != nil {
			return SourceAndMetadata{Source: source, Metadata: metadata}, fmt.Errorf("failed to update hierarchy: %w", err)
		}

		for _, child := range children {
			// it's a nested playlist which was already handled
			if child.Source.DurationMs == 0 {
				continue
			}
			// it's a track group which was already handled
			if len(child.Metadata.Chapters) > 0 {
				continue
			}

			track := extractTrackProperties(child.Source)
			track.ParentTitle = source.Title
			tracks = append(tracks, track)
		}
	} else if len(metadata.Chapters) > 0 {
		for _, chapter := range metadata.Chapters {
			track := extractTrackProperties(source)
			track.RawTitle = chapter.Title
			track.ParentTitle = source.Title
			track.StartOffsetMs = int64(chapter.StartTime * 1000)
			track.EndOffsetMs = int64(chapter.EndTime * 1000)
			tracks = append(tracks, track)
		}
	} else if metadata.Duration > 0 && parentId == uuid.Nil {
		// no parents left to handle this, we have to add it as a standalone track
		tracks = append(tracks, extractTrackProperties(source))
	}

	if len(tracks) > 0 {
		tracks, err = s.normalizer.Normalize(tracks)
		if err != nil {
			return SourceAndMetadata{Source: source, Metadata: metadata}, fmt.Errorf("failed to normalize tracks: %w", err)
		}

		tracksBySource := map[uuid.UUID][]storage.Track{}
		for _, trackProperties := range tracks {
			track := storage.Track{
				SourceId:      trackProperties.SourceId,
				Artist:        trackProperties.Artist,
				Title:         trackProperties.Title,
				StartOffsetMs: trackProperties.StartOffsetMs,
				EndOffsetMs:   trackProperties.EndOffsetMs,
			}

			if _, ok := tracksBySource[track.SourceId]; !ok {
				tracksBySource[track.SourceId] = []storage.Track{track}
			} else {
				tracksBySource[track.SourceId] = append(tracksBySource[track.SourceId], track)
			}
		}

		for sourceId, tracks := range tracksBySource {
			if _, err := s.initializeTracksFor(sourceId, tracks, managementPolicy); err != nil {
				return SourceAndMetadata{Source: source, Metadata: metadata}, fmt.Errorf("failed to initialize tracks for source %s: %w", sourceId, err)
			}
		}
	}

	return SourceAndMetadata{Source: source, Metadata: metadata}, nil
}

func parseDateOrNull(str string) *time.Time {
	result, err := time.Parse("20060102", str)
	if err != nil {
		return nil
	} else {
		return &result
	}
}

func extractTrackProperties(source storage.Source) TrackProperties {
	return TrackProperties{
		SourceId:      source.Id,
		RawTitle:      source.Title,
		Artist:        source.TrackArtist,
		Title:         source.TrackTitle,
		AlbumArtist:   source.AlbumArtist,
		Uploader:      source.Uploader,
		StartOffsetMs: 0,
		EndOffsetMs:   source.DurationMs,
	}
}

func (s *SourceService) initializeTracksFor(sourceId uuid.UUID, tracks []storage.Track, managementPolicy model.SourceManagementPolicy) ([]storage.Track, error) {
	currentManagementPolicy, err := s.storage.GetManagementPolicyById(sourceId)
	if err != nil {
		return tracks, fmt.Errorf("failed to get current source management policy: %w", err)
	}

	if currentManagementPolicy == model.SOURCE_MANAGEMENT_POLICY_MANUAL && managementPolicy != model.SOURCE_MANAGEMENT_POLICY_MANUAL {
		return s.tracks.GetDirectTracksBySource(sourceId)
	}

	if managementPolicy == model.SOURCE_MANAGEMENT_POLICY_MANUAL && currentManagementPolicy != managementPolicy {
		if err := s.storage.SetManagementPolicyById(sourceId, managementPolicy); err != nil {
			return tracks, fmt.Errorf("failed to update source management policy: %w", err)
		}
	}

	return s.tracks.InitializeTracksFor(sourceId, tracks)
}

func (s *SourceService) ReplaceTracksFor(sourceId uuid.UUID, tracks []storage.Track, managementPolicy model.SourceManagementPolicy) ([]storage.Track, error) {
	currentManagementPolicy, err := s.storage.GetManagementPolicyById(sourceId)
	if err != nil {
		return tracks, fmt.Errorf("failed to get current source management policy: %w", err)
	}

	if currentManagementPolicy == model.SOURCE_MANAGEMENT_POLICY_MANUAL && managementPolicy != model.SOURCE_MANAGEMENT_POLICY_MANUAL {
		return s.tracks.GetDirectTracksBySource(sourceId)
	}

	if managementPolicy == model.SOURCE_MANAGEMENT_POLICY_MANUAL && currentManagementPolicy != managementPolicy {
		if err := s.storage.SetManagementPolicyById(sourceId, managementPolicy); err != nil {
			return tracks, fmt.Errorf("failed to update source management policy: %w", err)
		}
	}

	return s.tracks.ReplaceBySource(sourceId, tracks)
}

type ListSourceForApi struct {
	Source storage.Source
	File   *storage.SourceFile
}

func (s *SourceService) GetListForApi(managementPolicies []model.SourceManagementPolicy) ([]ListSourceForApi, error) {
	sources, err := s.storage.GetListForApi(managementPolicies)
	if err != nil {
		return []ListSourceForApi{}, err
	}

	sourceIds := []uuid.UUID{}
	for _, source := range sources {
		sourceIds = append(sourceIds, source.Id)
	}

	files, err := s.files.FindBySourceIds(sourceIds)
	if err != nil {
		return []ListSourceForApi{}, err
	}

	fileLookup := map[uuid.UUID]storage.SourceFile{}
	for _, file := range files {
		fileLookup[file.SourceId] = file
	}

	result := []ListSourceForApi{}
	for _, source := range sources {
		dto := ListSourceForApi{Source: source}
		if file, ok := fileLookup[source.Id]; ok {
			dto.File = &file
		}
		result = append(result, dto)
	}

	return result, nil
}

func (s *SourceService) GetHierarchy(id uuid.UUID) ([]storage.SourceForHierarchy, error) {
	return s.storage.GetHierarchy(id)
}

func (s *SourceService) GetById(id uuid.UUID) (storage.Source, error) {
	return s.storage.GetById(id)
}

func (s *SourceService) FindByUrl(url string) (*storage.Source, error) {
	return s.storage.FindByUrl(url)
}
