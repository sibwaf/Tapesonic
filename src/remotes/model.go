package remotes

import (
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
)

type RemoteType = string

const (
	REMOTE_TYPE_SUBSONIC RemoteType = "subsonic"
)

type RemoteStatus = string

const (
	REMOTE_STATUS_OK                RemoteStatus = "OK"
	REMOTE_STATUS_NOT_AUTHENTICATED RemoteStatus = "NOT_AUTHENTICATED"
	REMOTE_STATUS_UNREACHABLE       RemoteStatus = "UNREACHABLE"
	REMOTE_STATUS_BAD_RESPONSE      RemoteStatus = "BAD_RESPONSE"
)

type Remote struct {
	Id uuid.UUID

	Name string
	Type RemoteType
	Url  string

	IsScrobbleReplicationEnabled bool
	IsExternalScrobblingEnabled  bool

	CreatedAt util.TimestampWrapper
}

type RemoteToUser struct {
	RemoteId uuid.UUID `gorm:"primaryKey"`
	Remote   Remote

	UserId uuid.UUID `gorm:"primaryKey"`
	User   users.User

	Credentials RemoteCredentials `gorm:"serializer:json"`
}

type RemoteSettings struct {
	Name string
	Type RemoteType
	Url  string

	IsScrobbleReplicationEnabled bool
	IsExternalScrobblingEnabled  bool
}

type RemoteCredentials struct {
	Username string
	Password string
}

type RemoteInfo struct {
	Name             string
	Version          string
	Status           RemoteStatus
	ErrorDescription string
	Username         string
}

// tasks

const (
	TASK_REMOTES_SYNC_SUBSONIC_LIBRARY = "REMOTES_SYNC_SUBSONIC_LIBRARY"
)

type SyncSubsonicLibraryTask struct {
	UserId   uuid.UUID
	RemoteId uuid.UUID
}
