package remotes

import (
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteAlbum struct {
	Id uuid.UUID

	RemoteId uuid.UUID
	AlbumId  string

	Title      string
	ArtworkId  string
	ArtistId   string
	AddedAt    util.TimestampWrapper
	ReleasedAt *util.TimestampWrapper

	SearchTitle string
}

type RemoteAlbumToUser struct {
	UserId  uuid.UUID
	SyncTag string
}

type RemoteAlbumStorage struct {
	db *gorm.DB
}

func newRemoteAlbumStorage(db *gorm.DB) *RemoteAlbumStorage {
	return &RemoteAlbumStorage{db: db}
}

func (store *RemoteAlbumStorage) Upsert(album RemoteAlbum, albumToUser RemoteAlbumToUser) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		type IdHolder struct {
			Id uuid.UUID
		}

		sql1 := `
			INSERT INTO remote_albums (id, remote_id, album_id, artwork_id, artist_id, title, added_at, released_at, search_title)
			VALUES (@id, @remoteId, @albumId, @artworkId, @artistId, @title, @addedAt, @releasedAt, @searchTitle)
			ON CONFLICT (remote_id, album_id) DO UPDATE
			SET artwork_id = excluded.artwork_id, artist_id = excluded.artist_id, title = excluded.title, added_at = excluded.added_at, released_at = excluded.released_at, search_title = excluded.search_title
			RETURNING id
		`
		params1 := map[string]any{
			"id":          album.Id,
			"remoteId":    album.RemoteId,
			"albumId":     album.AlbumId,
			"artworkId":   album.ArtworkId,
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
			INSERT INTO remote_albums_to_users (remote_album_id, user_id, sync_tag)
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
		DELETE FROM remote_albums_to_users
		WHERE user_id = @userId AND sync_tag != @syncTag
	`
	params := map[string]any{
		"userId":  userId,
		"syncTag": syncTag,
	}
	return store.db.Exec(sql, params).Error
}
