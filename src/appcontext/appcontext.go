package appcontext

import (
	"os"
	"path"
	configPkg "tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/http/lastfm"
	"tapesonic/http/listenbrainz"
	"tapesonic/http/subsonic/client"
	"tapesonic/logic"
	"tapesonic/storage"
	"tapesonic/tasks"
	"tapesonic/util"
	"tapesonic/ytdlp"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Context struct {
	Config *configPkg.TapesonicConfig

	TapeStorage                 *storage.TapeStorage
	SourceStorage               *storage.SourceStorage
	SourceFileStorage           *storage.SourceFileStorage
	TrackStorage                *storage.TrackStorage
	ThumbnailStorage            *storage.ThumbnailStorage
	PlaylistStorage             *storage.PlaylistStorage
	AlbumStorage                *storage.AlbumStorage
	TrackListensStorage         *storage.TrackListensStorage
	CachedMuxSongStorage        *storage.CachedMuxSongStorage
	CachedMuxAlbumStorage       *storage.CachedMuxAlbumStorage
	CachedMuxArtistStorage      *storage.CachedMuxArtistStorage
	MuxedSongListensStorage     *storage.MuxedSongListensStorage
	ListenbrainzPlaylistStorage *storage.ListenbrainzPlaylistStorage
	LastFmSessionStorage        *storage.LastFmSessionStorage
	YtdlpMetadataStorage        *storage.YtdlpMetadataStorage
	MediaStorage                *storage.MediaStorage
	StreamCacheStorage          *storage.StreamCacheStorage

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg

	YtdlpService *logic.YtdlpService

	ListenBrainzClient *listenbrainz.ListenBrainzClient
	LastFmClient       *lastfm.LastFmClient

	LastFmService *logic.LastFmService

	ThumbnailService *logic.ThumbnailService

	TrackNormalizer   *logic.TrackNormalizer
	TrackService      *logic.TrackService
	SourceFileService *logic.SourceFileService
	SourceService     *logic.SourceService
	TapeService       *logic.TapeService

	SearchService *logic.SearchService

	SubsonicProviders []*logic.SubsonicNamedService
	SubsonicMuxer     logic.SubsonicService
	SubsonicService   logic.SubsonicService

	ScrobbleService *logic.ScrobbleService
}

func NewContext(config *configPkg.TapesonicConfig) (*Context, error) {
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
		sqlite.Open(path.Join(config.DataStorageDir, "data.sqlite?_foreign_keys=on&_journal_mode=wal")),
		&gorm.Config{
			Logger: slogGorm.New(
				slogGorm.SetLogLevel(slogGorm.DefaultLogType, configPkg.LevelTrace),
				slogGorm.WithTraceAll(),
			),
		},
	)
	if err != nil {
		return nil, err
	}

	if context.TapeStorage, err = storage.NewTapeStorage(db); err != nil {
		return nil, err
	}
	if context.SourceStorage, err = storage.NewSourceStorage(db); err != nil {
		return nil, err
	}
	if context.SourceFileStorage, err = storage.NewSourceFileStorage(db); err != nil {
		return nil, err
	}
	if context.TrackStorage, err = storage.NewTrackStorage(db); err != nil {
		return nil, err
	}
	if context.ThumbnailStorage, err = storage.NewThumbnailStorage(db); err != nil {
		return nil, err
	}
	if context.PlaylistStorage, err = storage.NewPlaylistStorage(db); err != nil {
		return nil, err
	}
	if context.AlbumStorage, err = storage.NewAlbumStorage(db); err != nil {
		return nil, err
	}
	if context.TrackListensStorage, err = storage.NewTrackListensStorage(db); err != nil {
		return nil, err
	}
	if context.CachedMuxSongStorage, err = storage.NewCachedMuxSongStorage(db); err != nil {
		return nil, err
	}
	if context.CachedMuxAlbumStorage, err = storage.NewCachedMuxAlbumStorage(db); err != nil {
		return nil, err
	}
	if context.CachedMuxArtistStorage, err = storage.NewCachedMuxArtistStorage(db); err != nil {
		return nil, err
	}
	if context.MuxedSongListensStorage, err = storage.NewMuxedSongListensStorage(db); err != nil {
		return nil, err
	}
	if context.ListenbrainzPlaylistStorage, err = storage.NewListenBrainzPlaylistStorage(db); err != nil {
		return nil, err
	}
	if context.LastFmSessionStorage, err = storage.NewLastFmSessionStorage(db); err != nil {
		return nil, err
	}
	if context.YtdlpMetadataStorage, err = storage.NewYtdlpMetadataStorage(db); err != nil {
		return nil, err
	}

	if err = storage.Migrate(db); err != nil {
		return nil, err
	}

	context.MediaStorage = storage.NewMediaStorage(
		db,
		config.MediaStorageDir,
		context.TapeStorage,
		context.PlaylistStorage,
		context.AlbumStorage,
	)

	if context.StreamCacheStorage, err = storage.NewStreamCacheStorage(
		path.Join(config.CacheDir, "stream"),
		config.StreamCacheSize,
		config.StreamCacheMinLifetime,
		db,
	); err != nil {
		return nil, err
	}

	context.YtdlpService = logic.NewYtdlpService(
		context.Ytdlp,
		context.YtdlpMetadataStorage,
		config.YtdlpMetadataMaxLifetime,
		config.YtdlpMetadataMaxParallelism,
	)

	if config.ListenBrainzToken != "" {
		context.ListenBrainzClient = listenbrainz.NewListenBrainzClient(config.ListenBrainzToken)
	}

	if config.LastFmApiKey != "" && config.LastFmApiSecret != "" {
		context.LastFmClient = lastfm.NewLastFmClient(config.LastFmApiKey, config.LastFmApiSecret)
	}

	context.LastFmService = logic.NewLastFmService(
		context.LastFmClient,
		context.LastFmSessionStorage,
	)

	context.ScrobbleService = logic.NewScrobbleService(
		context.ListenBrainzClient,
		context.LastFmService,
	)

	context.ThumbnailService = logic.NewThumbnailService(
		context.ThumbnailStorage,
		path.Join(config.MediaStorageDir, "thumbnails"),
	)

	context.TrackNormalizer = logic.NewTrackNormalizer()
	context.TrackService = logic.NewTrackService(context.TrackStorage)
	context.SourceFileService = logic.NewSourceFileService(
		context.SourceFileStorage,
		context.SourceStorage,
		context.YtdlpService,
		config.MediaStorageDir,
	)
	context.SourceService = logic.NewSourceService(
		context.SourceStorage,
		context.YtdlpService,
		context.SourceFileService,
		context.TrackService,
		context.ThumbnailService,
		context.TrackNormalizer,
	)
	context.TapeService = logic.NewTapeService(context.TapeStorage, context.TrackStorage)

	context.SearchService = logic.NewSearchService(context.SourceStorage, context.TrackStorage)

	subsonicMux := logic.NewSubsonicMuxService(
		context.CachedMuxSongStorage,
		context.MuxedSongListensStorage,
		util.TakeIf(context.ScrobbleService, config.ScrobbleMode == configPkg.ScrobbleAll),
	)
	context.SubsonicMuxer = subsonicMux

	internalSubsonic := logic.NewSubsonicNamedService(
		"tapesonic",
		logic.NewSubsonicInternalService(
			context.TrackStorage,
			context.AlbumStorage,
			context.PlaylistStorage,
			context.TrackListensStorage,
			context.MediaStorage,
			context.StreamCacheStorage,
			context.Ffmpeg,
			context.ThumbnailService,
			util.TakeIf(context.ScrobbleService, config.ScrobbleMode == configPkg.ScrobbleTapesonic),
			context.YtdlpService,
		),
	)
	context.SubsonicProviders = append(context.SubsonicProviders, internalSubsonic)
	subsonicMux.AddService(internalSubsonic)

	if config.SubsonicProxyUrl != "" {
		externalSubsonic := logic.NewSubsonicNamedService(
			"proxy",
			logic.NewSubsonicExternalService(
				client.NewSubsonicClient(
					config.SubsonicProxyUrl,
					config.SubsonicProxyUsername,
					config.SubsonicProxyPassword,
				),
			),
		)
		context.SubsonicProviders = append(context.SubsonicProviders, externalSubsonic)
		subsonicMux.AddService(externalSubsonic)
	}

	context.SubsonicService = logic.NewSubsonicMainService(
		subsonicMux,
		context.SubsonicProviders,
		context.CachedMuxSongStorage,
		context.CachedMuxAlbumStorage,
		context.CachedMuxArtistStorage,
		context.ListenbrainzPlaylistStorage,
	)

	if err = registerBackgroundTasks(&context); err != nil {
		return nil, err
	}

	return &context, nil
}

func registerBackgroundTasks(context *Context) error {
	var err error

	cron := cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
		cron.WithSeconds(),
	)

	if context.Config.TasksDownloadSources.Cron != configPkg.CronDisabled {
		if err = tasks.NewDownloadSourcesTaskHandler(
			context.SourceFileService,
			context.SourceStorage,
			context.Config.TasksDownloadSources,
		).RegisterSchedules(cron); err != nil {
			return err
		}
	}

	if context.Config.TasksListenBrainzPlaylistSync.Cron != configPkg.CronDisabled && context.ListenBrainzClient != nil {
		if err = tasks.NewListenBrainzPlaylistSyncHandler(
			context.ListenBrainzClient,
			context.CachedMuxSongStorage,
			*context.ListenbrainzPlaylistStorage,
			context.Config.TasksListenBrainzPlaylistSync,
		).RegisterSchedules(cron); err != nil {
			return err
		}
	}

	if context.Config.TasksLibrarySync.Cron != configPkg.CronDisabled {
		if err = tasks.NewSyncLibraryHandler(
			context.SubsonicProviders,
			context.CachedMuxSongStorage,
			context.CachedMuxAlbumStorage,
			context.CachedMuxArtistStorage,
			context.Config.TasksLibrarySync,
		).RegisterSchedules(cron); err != nil {
			return err
		}
	}

	cron.Start()

	return nil
}
