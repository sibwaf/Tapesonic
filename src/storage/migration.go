package storage

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.Exec("UPDATE tape_files SET audio_codec = (metadata ->> 'acodec') WHERE audio_codec IS NULL").Error; err != nil {
		return err
	}
	if err := db.Exec("UPDATE tape_files SET audio_format = (metadata ->> 'audio_ext') WHERE audio_format IS NULL").Error; err != nil {
		return err
	}

	return nil
}
