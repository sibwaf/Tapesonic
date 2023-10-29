package appcontext

import (
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

type Context struct {
	Config *config.TapesonicConfig

	Storage  *storage.Storage
	Importer *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg
}

func NewContext(config *config.TapesonicConfig) *Context {
	context := Context{
		Config: config,

		Ytdlp:  ytdlp.NewYtdlp(config.YtdlpPath),
		Ffmpeg: ffmpeg.NewFfmpeg(config.FfmpegPath),

		Storage: storage.NewStorage(config.MediaStorageDir),
	}

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
	)

	return &context
}
