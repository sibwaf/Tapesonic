package lastfm

import (
	"errors"
	"tapesonic/util"

	"github.com/google/uuid"
)

// errors

var (
	ErrLastFmNotConfigured = errors.New("last.fm client is not configured")
	ErrNoActiveSession     = errors.New("no active last.fm session")
)

// internal

type LastFmAuthLink struct {
	Url   string
	Token string
}

type LastFmSession struct {
	UserId uuid.UUID

	SessionKey string
	Username   string

	IsScrobblingEnabled bool

	CreatedAt util.TimestampWrapper
	UpdatedAt util.TimestampWrapper
}

type SessionSettings struct {
	IsScrobblingEnabled bool
}

type Playlist struct {
	Id   string
	Name string
}

// http

type TokenWrapper struct {
	Token string `json:"token"`
}

type SessionWrapper struct {
	Session Session `json:"session"`
}

type Session struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type UpdateNowPlayingRq struct {
	Artist string
	Track  string
	Album  string
}

type ScrobbleRq struct {
	Artist    string
	Track     string
	Album     string
	Timestamp int64
}

type PlaylistWrapperRs struct {
	Items []PlaylistItemRs `json:"playlist"`
}

type PlaylistItemRs struct {
	Url  string `json:"url"`
	Name string `json:"name"`

	Artists   []ArtistRs   `json:"artists"`
	Playlinks []PlaylinkRs `json:"playlinks"`
}

type ArtistRs struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type PlaylinkRs struct {
	Affiliate string `json:"affiliate"`
	Url       string `json:"url"`
}
