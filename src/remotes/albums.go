package remotes

import (
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteAlbum struct {
	Id uuid.UUID

	RemoteId uuid.UUID `gorm:"uniqueIndex:remote_album_uniq"`
	Remote   Remote

	AlbumId string `gorm:"uniqueIndex:remote_album_uniq"`

	CoverId    string
	ArtistId   string
	Title      string
	AddedAt    util.TimestampWrapper
	ReleasedAt *util.TimestampWrapper

	SearchTitle string

	Users []RemoteAlbumToUser `gorm:"constraint:OnDelete:CASCADE;"`
}

type RemoteAlbumToUser struct {
	RemoteAlbumId uuid.UUID `gorm:"uniqueIndex:remote_album_to_user_uniq"`
	RemoteAlbum   RemoteAlbum

	UserId uuid.UUID `gorm:"uniqueIndex:remote_album_to_user_uniq"`
	User   users.User

	SyncTag string
}

type RemoteAlbumStorage struct {
	db *gorm.DB
}

func newRemoteAlbumStorage(db *gorm.DB) (*RemoteAlbumStorage, error) {
	if err := db.AutoMigrate(&RemoteAlbum{}, &RemoteAlbumToUser{}); err != nil {
		return nil, err
	}

	return &RemoteAlbumStorage{db: db}, nil
}

func (store *RemoteAlbumStorage) Upsert(album RemoteAlbum, albumToUser RemoteAlbumToUser) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		type IdHolder struct {
			Id uuid.UUID
		}

		sql1 := `
			INSERT INTO remote_albums (id, remote_id, album_id, cover_id, artist_id, title, added_at, released_at, search_title)
			VALUES (@id, @remoteId, @albumId, @coverId, @artistId, @title, @addedAt, @releasedAt, @searchTitle)
			ON CONFLICT (remote_id, album_id) DO UPDATE
			SET cover_id = excluded.cover_id, artist_id = excluded.artist_id, title = excluded.title, added_at = excluded.added_at, released_at = excluded.released_at, search_title = excluded.search_title
			RETURNING id
		`
		params1 := map[string]any{
			"id":          album.Id,
			"remoteId":    album.RemoteId,
			"albumId":     album.AlbumId,
			"coverId":     album.CoverId,
			"artistId":    album.ArtistId,
			"title":       album.Title,
			"addedAt":     album.AddedAt,
			"releasedAt":  album.ReleasedAt,
			"searchTitle": util.MakeTextSearchString(album.Title),
		}
		remoteAlbumIdHolder := IdHolder{}
		if err := tx.Raw(sql1, params1).First(&remoteAlbumIdHolder).Error; err != nil {
			return err
		}

		sql2 := `
			INSERT INTO remote_album_to_users (remote_album_id, user_id, sync_tag)
			VALUES (@remoteAlbumId, @userId, @syncTag)
			ON CONFLICT (remote_album_id, user_id) DO UPDATE
			SET sync_tag = excluded.sync_tag
		`
		params2 := map[string]any{
			"remoteAlbumId": remoteAlbumIdHolder.Id,
			"userId":        albumToUser.UserId,
			"syncTag":       albumToUser.SyncTag,
		}
		if err := tx.Exec(sql2, params2).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *RemoteAlbumStorage) DeleteUserBindingsBySyncTag(userId uuid.UUID, syncTag string) error {
	sql := `
		DELETE FROM remote_album_to_users
		WHERE user_id = @userId AND sync_tag != @syncTag
	`
	params := map[string]any{
		"userId":  userId,
		"syncTag": syncTag,
	}
	return store.db.Exec(sql, params).Error
}
