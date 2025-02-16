package responses

import (
	"tapesonic/storage"
	"time"

	"github.com/google/uuid"
)

type FullSourceRs struct {
	Id uuid.UUID

	Url      string
	Title    string
	Uploader string

	AlbumArtist string
	AlbumTitle  string
	AlbumIndex  int
	TrackArtist string
	TrackTitle  string
	DurationMs  int64

	ReleaseDate *time.Time

	ThumbnailId *uuid.UUID
}

func SourceToFullSourceRs(source storage.Source) FullSourceRs {
	return FullSourceRs{
		Id: source.Id,

		Url:      source.Url,
		Title:    source.Title,
		Uploader: source.Uploader,

		AlbumArtist: source.AlbumArtist,
		AlbumTitle:  source.AlbumTitle,
		AlbumIndex:  source.AlbumIndex,
		TrackArtist: source.TrackArtist,
		TrackTitle:  source.TrackTitle,
		DurationMs:  source.DurationMs,

		ReleaseDate: source.ReleaseDate,

		ThumbnailId: source.ThumbnailId,
	}
}

type ListSourceRs struct {
	Id uuid.UUID

	Url      string
	Title    string
	Uploader string

	DurationMs int64

	ThumbnailId *uuid.UUID
}

func SourcesToListSourceRs(sources []storage.Source) []ListSourceRs {
	sourceDtos := []ListSourceRs{}
	for _, source := range sources {
		sourceDtos = append(sourceDtos, SourceToListSourceRs(source))
	}
	return sourceDtos
}

func SourceToListSourceRs(source storage.Source) ListSourceRs {
	return ListSourceRs{
		Id: source.Id,

		Url:      source.Url,
		Title:    source.Title,
		Uploader: source.Uploader,

		DurationMs: source.DurationMs,

		ThumbnailId: source.ThumbnailId,
	}
}

type ListSourceHierarchyRs struct {
	Id       uuid.UUID
	ParentId *uuid.UUID

	Url      string
	Title    string
	Uploader string

	ListIndex int

	ThumbnailId *uuid.UUID
}

func SourcesForHierarchyToListSourceHierarchyRs(sources []storage.SourceForHierarchy) []ListSourceHierarchyRs {
	sourceDtos := []ListSourceHierarchyRs{}
	for _, source := range sources {
		sourceDtos = append(sourceDtos, SourceForHierarchyToListSourceHierarchyRs(source))
	}
	return sourceDtos
}

func SourceForHierarchyToListSourceHierarchyRs(source storage.SourceForHierarchy) ListSourceHierarchyRs {
	return ListSourceHierarchyRs{
		Id:       source.Id,
		ParentId: source.ParentId,

		Url:      source.Url,
		Title:    source.Title,
		Uploader: source.Uploader,

		ListIndex: source.ListIndex,

		ThumbnailId: source.ThumbnailId,
	}
}
