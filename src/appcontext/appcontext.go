package appcontext

import (
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

type Context struct {
	Config *config.TapesonicConfig

	DataStorage  *storage.DataStorage
	MediaStorage *storage.MediaStorage
	Importer     *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg
}

func NewContext(config *config.TapesonicConfig) (*Context, error) {
	context := Context{
		Config: config,

		Ytdlp:  ytdlp.NewYtdlp(config.YtdlpPath),
		Ffmpeg: ffmpeg.NewFfmpeg(config.FfmpegPath),
	}

	var err error
	context.DataStorage, err = storage.NewDataStorage(config.DataStorageDir)
	if err != nil {
		return nil, err
	}

	context.MediaStorage = storage.NewMediaStorage(
		config.MediaStorageDir,
		context.DataStorage,
	)

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
		context.DataStorage,
	)

	return &context, nil
}
