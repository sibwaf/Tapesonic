package sources

import (
	"tapesonic/artists"
	"tapesonic/storage"
	"tapesonic/util"
	"tapesonic/ytdlp"

	"github.com/google/uuid"
)

// internal

type SourceManagementPolicy = string

const (
	SOURCE_MANAGEMENT_POLICY_MANUAL SourceManagementPolicy = "MANUAL"
	SOURCE_MANAGEMENT_POLICY_AUTO   SourceManagementPolicy = "AUTO"
)

type SourceForApi struct {
	Source Source
	File   *SourceFile
}

type SourceForHierarchy struct {
	Id       uuid.UUID
	ParentId *uuid.UUID

	Url      string
	Title    string
	Uploader string

	ListIndex int

	ThumbnailId *uuid.UUID
}

type AnalyzedSourceTree struct {
	Metadata ytdlp.YtdlpFile
	Source   Source
	Children []AnalyzedSourceTree
	Tracks   []TrackProperties
}

type SourceTree struct {
	Source   Source
	Children []SourceTree
	Tracks   []SavedSourceTrack
}

func (node SourceTree) CollectTracks() []SavedSourceTrack {
	result := append([]SavedSourceTrack{}, node.Tracks...)
	for _, child := range node.Children {
		result = append(result, child.CollectTracks()...)
	}
	return result
}

type TrackProperties struct {
	SourceId uuid.UUID

	Uploader string

	RawTitle    string
	ParentTitle string

	RawAlbumArtist string
	RawTrackArtist string
	RawTrackTitle  string

	ArtistId *uuid.UUID

	Artist string
	Title  string

	StartOffsetMs int64
	EndOffsetMs   int64
}

type SavedSourceTrack struct {
	Id       uuid.UUID
	SourceId uuid.UUID

	ArtistId   *uuid.UUID
	ArtistName string

	Title string

	StartOffsetMs int64
	EndOffsetMs   int64
}

type SourceTrackForMetadataGuessing struct {
	Id uuid.UUID

	AlbumArtist string
	AlbumTitle  string

	SourceTitle        string
	SourceParentTitles []string `gorm:"serializer:json"`

	ArtistId    *uuid.UUID
	ReleaseDate *util.TimestampWrapper
	ThumbnailId *uuid.UUID
}

// database

type Source struct {
	Id uuid.UUID

	ExtractorKey string
	ExtractedId  string
	Url          string `gorm:"uniqueIndex"`

	Title      string
	Uploader   string
	UploaderId string

	AlbumArtist string
	AlbumTitle  string
	AlbumIndex  int
	TrackArtist string
	TrackTitle  string
	DurationMs  int64

	UploadedAt  util.TimestampWrapper
	ReleaseDate *util.TimestampWrapper

	ThumbnailId *uuid.UUID
	Thumbnail   *storage.Thumbnail

	ManagementPolicy SourceManagementPolicy

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type SourceHierarchy struct {
	ParentId uuid.UUID `gorm:"primaryKey"`
	Parent   Source

	ChildId uuid.UUID `gorm:"primaryKey"`
	Child   Source

	ListIndex int
}

type SourceTrack struct {
	Id uuid.UUID

	SourceId uuid.UUID
	Source   Source

	StartOffsetMs int64
	EndOffsetMs   int64

	ArtistId *uuid.UUID
	Artist   *artists.Artist

	Title string

	SearchTitle string
}

type SourceFile struct {
	Id uuid.UUID

	SourceId uuid.UUID `gorm:"uniqueIndex"`
	Source   Source

	Format string
	Codec  string

	MediaPath string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

// tasks

const (
	TASK_SOURCES_FIND_SOURCE_FOR_DOWNLOAD = "SOURCES_FIND_SOURCE_FOR_DOWNLOAD"
)
