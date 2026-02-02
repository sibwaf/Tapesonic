package sources

import (
	"context"
	"fmt"
	"tapesonic/logic"
	"tapesonic/storage"
	"tapesonic/util"
	"tapesonic/ytdlp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type ImportService struct {
	sources    *SourceStorage
	tracks     *TrackStorage
	normalizer *TrackNormalizer
	ytdlp      *logic.YtdlpService
	thumbnails *logic.ThumbnailService
}

func newImportService(
	sources *SourceStorage,
	tracks *TrackStorage,
	normalizer *TrackNormalizer,
	ytdlp *logic.YtdlpService,
	thumbnails *logic.ThumbnailService,
) *ImportService {
	return &ImportService{
		sources:    sources,
		tracks:     tracks,
		normalizer: normalizer,
		ytdlp:      ytdlp,
		thumbnails: thumbnails,
	}
}

func (s *ImportService) AddSource(ctx context.Context, url string, managementPolicy SourceManagementPolicy) (Source, error) {
	result, err := s.addSourceRecursive(ctx, url, managementPolicy, uuid.Nil)
	return result.Source, err
}

type SourceAndMetadata struct {
	Source   Source
	Metadata ytdlp.YtdlpFile
}

func (s *ImportService) addSourceRecursive(ctx context.Context, url string, managementPolicy SourceManagementPolicy, parentId uuid.UUID) (SourceAndMetadata, error) {
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
	existingSource, err := s.sources.FindByUrl(metadata.WebpageUrl)
	if err != nil {
		return SourceAndMetadata{}, err
	}

	source := Source{
		Id: uuid.New(),

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

		UploadedAt:  util.NewTimestampWrapper(time.Unix(int64(metadata.Timestamp), 0)),
		ReleaseDate: util.NewTimestampWrapperOrNull(parseDateOrNull(metadata.ReleaseDate)),

		CreatedAt: util.NewTimestampWrapper(time.Now()),
		UpdatedAt: util.NewTimestampWrapper(time.Now()),
	}

	if thumbnail != nil {
		source.ThumbnailId = &thumbnail.Id
	}

	// never override MANUAL management policy
	if existingSource == nil || existingSource.ManagementPolicy != SOURCE_MANAGEMENT_POLICY_MANUAL {
		source.ManagementPolicy = managementPolicy
	} else {
		source.ManagementPolicy = existingSource.ManagementPolicy
	}

	source, err = s.sources.Upsert(source)
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

		if err := s.sources.UpdateHierarchy(source.Id, childIds); err != nil {
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

		tracksBySource := map[uuid.UUID][]SourceTrack{}
		for _, trackProperties := range tracks {
			track := SourceTrack{
				SourceId:      trackProperties.SourceId,
				Artist:        trackProperties.Artist,
				Title:         trackProperties.Title,
				StartOffsetMs: trackProperties.StartOffsetMs,
				EndOffsetMs:   trackProperties.EndOffsetMs,
			}

			if _, ok := tracksBySource[track.SourceId]; !ok {
				tracksBySource[track.SourceId] = []SourceTrack{track}
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

func extractTrackProperties(source Source) TrackProperties {
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

func (s *ImportService) initializeTracksFor(sourceId uuid.UUID, tracks []SourceTrack, managementPolicy SourceManagementPolicy) ([]SourceTrack, error) {
	currentManagementPolicy, err := s.sources.GetManagementPolicyById(sourceId)
	if err != nil {
		return tracks, err
	}

	existingTracks, err := s.tracks.GetDirectTracksBySource(sourceId)
	if err != nil {
		return tracks, err
	}

	if currentManagementPolicy == SOURCE_MANAGEMENT_POLICY_MANUAL && managementPolicy != SOURCE_MANAGEMENT_POLICY_MANUAL {
		return existingTracks, nil
	}

	if len(existingTracks) > 0 {
		return existingTracks, nil
	} else {
		for i := range tracks {
			tracks[i].Id = uuid.New()
		}

		return s.tracks.ReplaceTracksForSource(sourceId, tracks)
	}
}
