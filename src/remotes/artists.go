package remotes

import (
	"tapesonic/users"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RemoteArtist struct {
	Id uuid.UUID

	RemoteId uuid.UUID `gorm:"uniqueIndex:remote_artist_uniq"`
	Remote   Remote

	ArtistId string `gorm:"uniqueIndex:remote_artist_uniq"`

	CoverId string

	Name string

	SearchName string

	Users []RemoteArtistToUser `gorm:"constraint:OnDelete:CASCADE;"`
}

type RemoteArtistToUser struct {
	RemoteArtistId uuid.UUID `gorm:"uniqueIndex:remote_artist_to_user_uniq"`
	RemoteArtist   RemoteArtist

	UserId uuid.UUID `gorm:"uniqueIndex:remote_artist_to_user_uniq"`
	User   users.User

	SyncTag string
}

type RemoteArtistStorage struct {
	db *gorm.DB
}

func newRemoteArtistStorage(db *gorm.DB) (*RemoteArtistStorage, error) {
	if err := db.AutoMigrate(&RemoteArtist{}, &RemoteArtistToUser{}); err != nil {
		return nil, err
	}

	return &RemoteArtistStorage{db: db}, nil
}

func (store *RemoteArtistStorage) Upsert(artist RemoteArtist, artistToUser RemoteArtistToUser) error {
	return store.db.Transaction(func(tx *gorm.DB) error {
		type IdHolder struct {
			Id uuid.UUID
		}

		sql1 := `
			INSERT INTO remote_artists (id, remote_id, artist_id, cover_id, name, search_name)
			VALUES (@id, @remoteId, @artistId, @coverId, @name, @searchName)
			ON CONFLICT (remote_id, artist_id) DO UPDATE
			SET cover_id = excluded.cover_id, name = excluded.name, search_name = excluded.search_name
			RETURNING id
		`
		params1 := map[string]any{
			"id":         artist.Id,
			"remoteId":   artist.RemoteId,
			"artistId":   artist.ArtistId,
			"coverId":    artist.CoverId,
			"name":       artist.Name,
			"searchName": util.MakeTextSearchString(artist.Name),
		}
		remoteArtistIdHolder := IdHolder{}
		if err := tx.Raw(sql1, params1).First(&remoteArtistIdHolder).Error; err != nil {
			return err
		}

		sql2 := `
			INSERT INTO remote_artist_to_users (remote_artist_id, user_id, sync_tag)
			VALUES (@remoteArtistId, @userId, @syncTag)
			ON CONFLICT (remote_artist_id, user_id) DO UPDATE
			SET sync_tag = excluded.sync_tag
		`
		params2 := map[string]any{
			"remoteArtistId": remoteArtistIdHolder.Id,
			"userId":         artistToUser.UserId,
			"syncTag":        artistToUser.SyncTag,
		}
		if err := tx.Exec(sql2, params2).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *RemoteArtistStorage) DeleteUserBindingsBySyncTag(userId uuid.UUID, syncTag string) error {
	sql := `
		DELETE FROM remote_artist_to_users
		WHERE user_id = @userId AND sync_tag != @syncTag
	`
	params := map[string]any{
		"userId":  userId,
		"syncTag": syncTag,
	}
	return store.db.Exec(sql, params).Error
}
