package sources

import (
	"context"
	"tapesonic/logic"
	"tapesonic/util"
	"tapesonic/ytdlp"
	"time"

	"github.com/google/uuid"
)

type ImportService struct {
	sources    *SourceStorage
	tracks     *TrackStorage
	ytdlp      *ytdlp.YtdlpService
	thumbnails *logic.ThumbnailService
}

func newImportService(
	sources *SourceStorage,
	tracks *TrackStorage,
	ytdlp *ytdlp.YtdlpService,
	thumbnails *logic.ThumbnailService,
) *ImportService {
	return &ImportService{
		sources:    sources,
		tracks:     tracks,
		ytdlp:      ytdlp,
		thumbnails: thumbnails,
	}
}

func (svc *ImportService) Analyze(ctx context.Context, url string) (AnalyzedSourceTree, error) {
	metadata, err := svc.ytdlp.GetMetadata(ctx, url)
	if err != nil {
		return AnalyzedSourceTree{}, err
	}

	source := Source{
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
	}

	children, err := util.ParallelMapContext(ctx, metadata.Entries, func(ctx context.Context, entry ytdlp.YtdlpFile) (AnalyzedSourceTree, error) {
		entryUrl := util.Coalesce(entry.WebpageUrl, entry.Url)

		child, err := svc.Analyze(ctx, entryUrl)
		if err != nil {
			return AnalyzedSourceTree{}, err
		}

		for i, track := range child.Tracks {
			child.Tracks[i].ParentTitle = util.Coalesce(track.ParentTitle, source.Title)
		}

		return child, nil
	})
	if err != nil {
		return AnalyzedSourceTree{}, err
	}

	tracks := []TrackProperties{}
	if len(metadata.Chapters) > 0 {
		for _, chapter := range metadata.Chapters {
			track := extractTrackProperties(source)
			track.RawTitle = chapter.Title
			track.ParentTitle = source.Title
			track.StartOffsetMs = int64(chapter.StartTime * 1000)
			track.EndOffsetMs = int64(chapter.EndTime * 1000)
			tracks = append(tracks, track)
		}
	} else if metadata.Duration > 0 {
		tracks = append(tracks, extractTrackProperties(source))
	}

	return AnalyzedSourceTree{
		Metadata: metadata,
		Source:   source,
		Children: children,
		Tracks:   tracks,
	}, nil
}

func (svc *ImportService) ImportTree(node AnalyzedSourceTree, managementPolicy SourceManagementPolicy) (SourceTree, error) {
	children := []SourceTree{}
	for _, child := range node.Children {
		childTree, err := svc.ImportTree(child, managementPolicy)
		if err != nil {
			return SourceTree{}, err
		}

		children = append(children, childTree)
	}

	existingSource, err := svc.sources.FindByUrl(node.Source.Url)
	if err != nil {
		return SourceTree{}, err
	}

	source := node.Source
	if existingSource == nil || managementPolicy == SOURCE_MANAGEMENT_POLICY_MANUAL || existingSource.ManagementPolicy != SOURCE_MANAGEMENT_POLICY_MANUAL {
		source.Id = uuid.New()
		source.ManagementPolicy = managementPolicy
		source.CreatedAt = util.NewTimestampWrapper(time.Now())
		source.UpdatedAt = util.NewTimestampWrapper(time.Now())

		if node.Metadata.Thumbnail != "" {
			thumbnail, err := svc.thumbnails.CreateFromUrl(node.Metadata.Thumbnail)
			if err != nil {
				return SourceTree{}, err
			}

			source.ThumbnailId = &thumbnail.Id
		}

		source, err = svc.sources.Upsert(source)
		if err != nil {
			return SourceTree{}, err
		}
	} else {
		source = *existingSource
	}

	if len(children) > 0 {
		err := svc.sources.UpdateHierarchy(
			source.Id,
			util.Map(children, func(node SourceTree) uuid.UUID { return node.Source.Id }),
		)
		if err != nil {
			return SourceTree{}, err
		}
	}

	tracks, err := svc.initializeTracksFor(
		source.Id,
		util.Map(node.Tracks, func(track TrackProperties) SourceTrack {
			return SourceTrack{
				Id:            uuid.New(),
				ArtistId:      track.ArtistId,
				Title:         track.Title,
				StartOffsetMs: track.StartOffsetMs,
				EndOffsetMs:   track.EndOffsetMs,
			}
		}),
	)
	if err != nil {
		return SourceTree{}, err
	}

	return SourceTree{
		Source:   source,
		Children: children,
		Tracks:   tracks,
	}, nil
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
		SourceId:       source.Id,
		Uploader:       source.Uploader,
		RawTitle:       source.Title,
		RawAlbumArtist: source.AlbumArtist,
		RawTrackArtist: source.TrackArtist,
		RawTrackTitle:  source.TrackTitle,
		StartOffsetMs:  0,
		EndOffsetMs:    source.DurationMs,
	}
}

func (svc *ImportService) initializeTracksFor(sourceId uuid.UUID, tracks []SourceTrack) ([]SavedSourceTrack, error) {
	existingTracks, err := svc.tracks.GetDirectTracksBySource(sourceId)
	if err != nil {
		return []SavedSourceTrack{}, err
	}

	if len(existingTracks) > 0 {
		return existingTracks, nil
	} else {
		return svc.tracks.ReplaceTracksForSource(sourceId, tracks)
	}
}
