package recommendations

import (
	"tapesonic/library"
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
)

type RecommendationProvider string

const (
	RecommendationProviderListenBrainz = "listenbrainz"
	RecommendationProviderLastFm       = "lastfm"
)

// database

type RecommendedPlaylist struct {
	Id uuid.UUID

	Provider           RecommendationProvider `gorm:"uniqueIndex:recommended_playlist_uniq"`
	ProviderPlaylistId string                 `gorm:"uniqueIndex:recommended_playlist_uniq"`

	UserId uuid.UUID
	User   users.User

	Name string

	CoverId *string

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper

	SyncTag string

	Tracks []RecommendedPlaylistTrack `gorm:"constraint:OnDelete:CASCADE;"`
}

type RecommendedPlaylistTrack struct {
	RecommendedPlaylistId uuid.UUID
	RecommendedPlaylist   RecommendedPlaylist

	Artist string
	Title  string

	TrackId string
	Track   library.AllTrackId

	TrackIndex int
}

// tasks

const (
	TASK_RECOMMENDATIONS_DISCOVER_LASTFM_SESSIONS        = "RECOMMENDATIONS_DISCOVER_LASTFM_SESSIONS"
	TASK_RECOMMENDATIONS_DISCOVER_LASTFM_PLAYLISTS       = "RECOMMENDATIONS_DISCOVER_LASTFM_PLAYLISTS"
	TASK_RECOMMENDATIONS_SYNC_LASTFM_PLAYLIST            = "RECOMMENDATIONS_SYNC_LASTFM_PLAYLIST"
	TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_SESSIONS  = "RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_SESSIONS"
	TASK_RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_PLAYLISTS = "RECOMMENDATIONS_DISCOVER_LISTENBRAINZ_PLAYLISTS"
	TASK_RECOMMENDATIONS_SYNC_LISTENBRAINZ_PLAYLIST      = "RECOMMENDATIONS_SYNC_LISTENBRAINZ_PLAYLIST"
)

type DiscoverListenBrainzRecommendedPlaylistsTask struct {
	UserId uuid.UUID
}

type SyncListenBrainzRecommendedPlaylistTask struct {
	PlaylistId uuid.UUID
}

type DiscoverLastFmRecommendedPlaylistsTask struct {
	UserId uuid.UUID
}

type SyncLastFmRecommendedPlaylistTask struct {
	PlaylistId uuid.UUID
}
