package remotes

import (
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteTrack struct {
	Id uuid.UUID

	RemoteId uuid.UUID `gorm:"uniqueIndex:remote_track_uniq"`
	Remote   Remote

	TrackId string `gorm:"uniqueIndex:remote_track_uniq"`

	Artist   string
	ArtistId string

	Album   string
	AlbumId string

	AlbumIndex int

	CoverId string

	Title string

	DurationMs int

	SearchTitle string

	Users []RemoteTrackToUser `gorm:"constraint:OnDelete:CASCADE;"`
}

type RemoteTrackToUser struct {
	RemoteTrackId uuid.UUID `gorm:"uniqueIndex:remote_track_to_user_uniq"`
	RemoteTrack   RemoteTrack

	UserId uuid.UUID `gorm:"uniqueIndex:remote_track_to_user_uniq"`
	User   users.User

	SyncTag string
}

type RemoteTrackStorage struct {
	db *gorm.DB
}

func newRemoteTrackStorage(db *gorm.DB) (*RemoteTrackStorage, error) {
	if err := db.AutoMigrate(&RemoteTrack{}, &RemoteTrackToUser{}); err != nil {
		return nil, err
	}

	return &RemoteTrackStorage{db: db}, nil
}

func (store *RemoteTrackStorage) Upsert(track RemoteTrack, trackToUser RemoteTrackToUser) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		type IdHolder struct {
			Id uuid.UUID
		}

		sql1 := `
			INSERT INTO remote_tracks (id, remote_id, track_id, artist, artist_id, album, album_id, album_index, cover_id, title, duration_ms, search_title)
			VALUES (@id, @remoteId, @trackId, @artist, @artistId, @album, @albumId, @albumIndex, @coverId, @title, @durationMs, @searchTitle)
			ON CONFLICT (remote_id, track_id) DO UPDATE
			SET artist = excluded.artist, artist_id = excluded.artist_id, album = excluded.album, album_id = excluded.album_id, album_index = excluded.album_index, cover_id = excluded.cover_id, title = excluded.title, duration_ms = excluded.duration_ms, search_title = excluded.search_title
			RETURNING id
		`
		params1 := map[string]any{
			"id":          track.Id,
			"remoteId":    track.RemoteId,
			"trackId":     track.TrackId,
			"artist":      track.Artist,
			"artistId":    track.ArtistId,
			"album":       track.Album,
			"albumId":     track.AlbumId,
			"albumIndex":  track.AlbumIndex,
			"coverId":     track.CoverId,
			"title":       track.Title,
			"durationMs":  track.DurationMs,
			"searchTitle": util.MakeTextSearchString(track.Title),
		}
		remoteTrackIdHolder := IdHolder{}
		if err := tx.Raw(sql1, params1).First(&remoteTrackIdHolder).Error; err != nil {
			return err
		}

		sql2 := `
			INSERT INTO remote_track_to_users (remote_track_id, user_id, sync_tag)
			VALUES (@remoteTrackId, @userId, @syncTag)
			ON CONFLICT (remote_track_id, user_id) DO UPDATE
			SET sync_tag = excluded.sync_tag
		`
		params2 := map[string]any{
			"remoteTrackId": remoteTrackIdHolder.Id,
			"userId":        trackToUser.UserId,
			"syncTag":       trackToUser.SyncTag,
		}
		if err := tx.Exec(sql2, params2).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *RemoteTrackStorage) DeleteUserBindingsBySyncTag(userId uuid.UUID, syncTag string) error {
	sql := `
		DELETE FROM remote_track_to_users
		WHERE user_id = @userId AND sync_tag != @syncTag
	`
	params := map[string]any{
		"userId":  userId,
		"syncTag": syncTag,
	}
	return store.db.Exec(sql, params).Error
}
