package scrobbling

import (
	"tapesonic/library"
	"tapesonic/users"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListenStat struct {
	UserId uuid.UUID `gorm:"uniqueIndex:listen_stat_uniq"`
	User   users.User

	TrackId string `gorm:"uniqueIndex:listen_stat_uniq"`
	Track   library.AllTrackId

	ListenCount    int
	LastListenedAt util.TimestampWrapper
}

type ListenStatStorage struct {
	db *gorm.DB
}

func newListenStatStorage(db *gorm.DB) *ListenStatStorage {
	return &ListenStatStorage{db: db}
}

func (store *ListenStatStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&ListenStat{})
}

func (store *ListenStatStorage) AddListen(userId uuid.UUID, trackId string, listenedAt time.Time) error {
	sql := `
		INSERT INTO listen_stats (user_id, track_id, listen_count, last_listened_at)
		VALUES (@userId, @trackId, 1, @listenedAt)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET listen_count = listen_count + 1, last_listened_at = max(last_listened_at, excluded.last_listened_at)
	`
	params := map[string]any{
		"userId":     userId,
		"trackId":    trackId,
		"listenedAt": util.NewTimestampWrapper(listenedAt),
	}

	return store.db.Exec(sql, params).Error
}

func (store *ListenStatStorage) TouchListenedAt(userId uuid.UUID, trackId string, listenedAt time.Time) error {
	sql := `
		INSERT INTO listen_stats (user_id, track_id, listen_count, last_listened_at)
		VALUES (@userId, @trackId, 0, @listenedAt)
		ON CONFLICT (user_id, track_id) DO UPDATE
		SET last_listened_at = max(last_listened_at, excluded.last_listened_at)
	`
	params := map[string]any{
		"userId":     userId,
		"trackId":    trackId,
		"listenedAt": util.NewTimestampWrapper(listenedAt),
	}

	return store.db.Exec(sql, params).Error
}
