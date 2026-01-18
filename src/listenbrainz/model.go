package listenbrainz

import (
	"errors"
	"tapesonic/users"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

// errors

var (
	ErrNoActiveSession = errors.New("no active ListenBrainz session")
)

// internal

type ListenBrainzSession struct {
	UserId uuid.UUID `gorm:"primaryKey"`
	User   users.User

	Token    string
	Username string

	IsScrobblingEnabled bool

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type SessionSettings struct {
	IsScrobblingEnabled bool
}

// http

const (
	ListenTypeSingle     = "single"
	ListenTypePlayingNow = "playing_now"
)

type SubmitListensRequest struct {
	ListenType string                            `json:"listen_type"`
	Payload    []SubmitListensRequestPayloadItem `json:"payload"`
}

type SubmitListensRequestPayloadItem struct {
	ListenedAt    int64                                        `json:"listened_at,omitempty"`
	TrackMetadata SubmitListensRequestPayloadItemTrackMetadata `json:"track_metadata"`
}

type SubmitListensRequestPayloadItemTrackMetadata struct {
	ArtistName  string `json:"artist_name"`
	TrackName   string `json:"track_name"`
	ReleaseName string `json:"release_name,omitempty"`
}

type ValidateTokenResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"user_name"`
}

type PlaylistsResponse struct {
	Playlists []PlaylistResponseWrapper `json:"playlists"`
}

type PlaylistResponseWrapper struct {
	Playlist PlaylistResponse `json:"playlist"`
}

type PlaylistResponse struct {
	Identifier string `json:"identifier"`

	Title      string    `json:"title"`
	Annotation string    `json:"annotation"`
	Creator    string    `json:"creator"`
	Date       time.Time `json:"date"`

	Track []PlaylistTrackResponse `json:"track"`
}

type PlaylistTrackResponse struct {
	Identifier []string `json:"identifier"`

	Creator string `json:"creator"`
	Album   string `json:"album"`
	Title   string `json:"title"`

	Duration int `json:"duration"`
}
