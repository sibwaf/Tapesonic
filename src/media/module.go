package media

import (
	"tapesonic/artworks"
	"tapesonic/ffmpeg"
	"tapesonic/remotes"
	"tapesonic/sources"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

type MediaModule struct {
	Artworks *ArtworkService
	Streams  *StreamService
}

func NewMediaModule(
	remotes *remotes.RemoteService,
	artworks *artworks.ArtworkService,
	sources *sources.SourceService,
	cache *storage.StreamCacheStorage,
	ffmpeg *ffmpeg.Ffmpeg,
	ytdlp *ytdlp.YtdlpService,
) *MediaModule {
	return &MediaModule{
		Artworks: newArtworkService(remotes, artworks),
		Streams:  newStreamService(remotes, sources, cache, ffmpeg, ytdlp),
	}
}
