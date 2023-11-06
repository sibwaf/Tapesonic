package appcontext

import (
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

type Context struct {
	Config *config.TapesonicConfig

	DataStorage *storage.DataStorage
	Storage     *storage.Storage
	Importer    *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg
}

func NewContext(config *config.TapesonicConfig) (*Context, error) {
	context := Context{
		Config: config,

		Ytdlp:  ytdlp.NewYtdlp(config.YtdlpPath),
		Ffmpeg: ffmpeg.NewFfmpeg(config.FfmpegPath),

		Storage: storage.NewStorage(config.MediaStorageDir),
	}

	var err error
	context.DataStorage, err = storage.NewDataStorage(config.DataStorageDir)
	if err != nil {
		return nil, err
	}

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
		context.DataStorage,
	)

	return &context, nil
}
