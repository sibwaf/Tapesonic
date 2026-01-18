package media

import (
	"tapesonic/ffmpeg"
	"tapesonic/logic"
	"tapesonic/remotes"
	"tapesonic/storage"
)

type MediaModule struct {
	Covers  *CoverService
	Streams *StreamService
}

func NewMediaModule(
	remotes *remotes.RemoteService,
	thumbnails *logic.ThumbnailService,
	media *storage.MediaStorage,
	cache *storage.StreamCacheStorage,
	ffmpeg *ffmpeg.Ffmpeg,
	ytdlp *logic.YtdlpService,
) *MediaModule {
	return &MediaModule{
		Covers:  newCoverService(remotes, thumbnails),
		Streams: newStreamService(remotes, media, cache, ffmpeg, ytdlp),
	}
}
