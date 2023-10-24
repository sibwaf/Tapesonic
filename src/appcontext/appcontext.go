package appcontext

import (
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

type Context struct {
	Config *config.TapesonicConfig

	Storage *storage.Storage

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg
}
