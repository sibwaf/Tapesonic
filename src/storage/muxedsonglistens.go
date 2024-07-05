package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MuxedSongListensStorage struct {
	db *DbHelper
}

type MuxedSongListens struct {
	ServiceName string `gorm:"primaryKey"`
	SongId      string `gorm:"primaryKey"`

	ListenCount int

	LastListenedAt time.Time
}

func NewMuxedSongListensStorage(db *gorm.DB) (*MuxedSongListensStorage, error) {
	err := db.AutoMigrate(
		&MuxedSongListens{},
	)
	return &MuxedSongListensStorage{db: NewDbHelper(db)}, err
}

func (storage *MuxedSongListensStorage) Record(serviceName string, songId string, listenedAt time.Time, incrementListenCount bool) error {
	return storage.db.ExclusiveTransaction(func(tx *gorm.DB) error {
		item := MuxedSongListens{}
		if err := tx.Where(&MuxedSongListens{ServiceName: serviceName, SongId: songId}).Find(&item).Error; err != nil {
			return err
		}

		item.ServiceName = serviceName
		item.SongId = songId
		if listenedAt.After(item.LastListenedAt) {
			if incrementListenCount {
				item.ListenCount += 1
			}
			item.LastListenedAt = listenedAt
		}

		return tx.Save(&item).Error
	})
}

func (storage *MuxedSongListensStorage) GetRecentAlbumListenStats(count int, offset int) ([]CachedAlbumId, error) {
	return storage.getAlbumListenStats(count, offset, "album_info.last_listened_at IS NOT NULL", "album_info.last_listened_at DESC")
}

func (storage *MuxedSongListensStorage) GetFrequentAlbumListenStats(count int, offset int) ([]CachedAlbumId, error) {
	return storage.getAlbumListenStats(count, offset, "album_info.total_play_time > 0", "album_info.total_play_time DESC")
}

func (storage *MuxedSongListensStorage) getAlbumListenStats(count int, offset int, filter string, order string) ([]CachedAlbumId, error) {
	query := `
		WITH album_info AS (
			SELECT
				cached_mux_songs.service_name AS service_name,
				cached_mux_songs.album_id AS album_id,
				max(muxed_song_listens.last_listened_at) AS last_listened_at,
				sum(muxed_song_listens.listen_count * cached_mux_songs.duration_sec) AS total_play_time
			FROM muxed_song_listens
			JOIN cached_mux_songs ON cached_mux_songs.service_name = muxed_song_listens.service_name AND cached_mux_songs.song_id = muxed_song_listens.song_id
			WHERE cached_mux_songs.album_id != '' AND cached_mux_songs.album_id IS NOT NULL
			GROUP BY cached_mux_songs.service_name, cached_mux_songs.album_id
		)
		SELECT
			album_info.service_name AS service_name,
			album_info.album_id AS id
		FROM album_info
	`

	if filter != "" {
		query += fmt.Sprintf(" WHERE %s", filter)
	}

	query += fmt.Sprintf(" ORDER BY %s", order)
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", count, offset)

	result := []CachedAlbumId{}
	return result, storage.db.Raw(query).Find(&result).Error
}
