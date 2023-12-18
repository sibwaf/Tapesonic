package appcontext

import (
	"log/slog"
	"os"
	"path"
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Context struct {
	Config *config.TapesonicConfig

	TapeStorage     *storage.TapeStorage
	PlaylistStorage *storage.PlaylistStorage
	MediaStorage    *storage.MediaStorage
	Importer        *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg
}

func NewContext(config *config.TapesonicConfig) (*Context, error) {
	var err error
	context := Context{
		Config: config,

		Ytdlp:  ytdlp.NewYtdlp(config.YtdlpPath),
		Ffmpeg: ffmpeg.NewFfmpeg(config.FfmpegPath),
	}

	if err := os.MkdirAll(config.DataStorageDir, 0777); err != nil {
		return nil, err
	}
	db, err := gorm.Open(
		sqlite.Open(path.Join(config.DataStorageDir, "data.sqlite?_foreign_keys=on")),
		&gorm.Config{
			Logger: slogGorm.New(
				slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelDebug),
				slogGorm.WithTraceAll(),
			),
		},
	)
	if err != nil {
		return nil, err
	}

	context.TapeStorage, err = storage.NewTapeStorage(db)
	if err != nil {
		return nil, err
	}
	context.PlaylistStorage, err = storage.NewPlaylistStorage(db)
	if err != nil {
		return nil, err
	}

	context.MediaStorage = storage.NewMediaStorage(
		config.MediaStorageDir,
		context.TapeStorage,
		context.PlaylistStorage,
	)

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
		context.TapeStorage,
	)

	return &context, nil
}
