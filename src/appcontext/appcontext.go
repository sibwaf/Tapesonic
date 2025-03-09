package appcontext

import (
	"context"
	"fmt"
	"log/slog"
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
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Context struct {
	Config *configPkg.TapesonicConfig

	TapeStorage             *storage.TapeStorage
	SourceStorage           *storage.SourceStorage
	SourceFileStorage       *storage.SourceFileStorage
	TrackStorage            *storage.TrackStorage
	ThumbnailStorage        *storage.ThumbnailStorage
	PlaylistStorage         *storage.PlaylistStorage
	AlbumStorage            *storage.AlbumStorage
	TrackListensStorage     *storage.TrackListensStorage
	CachedMuxSongStorage    *storage.CachedMuxSongStorage
	CachedMuxAlbumStorage   *storage.CachedMuxAlbumStorage
	CachedMuxArtistStorage  *storage.CachedMuxArtistStorage
	MuxedSongListensStorage *storage.MuxedSongListensStorage
	ExternalPlaylistStorage *storage.ExternalPlaylistStorage
	LastFmSessionStorage    *storage.LastFmSessionStorage
	YtdlpMetadataStorage    *storage.YtdlpMetadataStorage
	MediaStorage            *storage.MediaStorage
	StreamCacheStorage      *storage.StreamCacheStorage

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg

	YtdlpService *logic.YtdlpService

	ListenBrainzClient *listenbrainz.ListenBrainzClient
	LastFmClient       *lastfm.LastFmClient

	LastFmService *logic.LastFmService

	ThumbnailService *logic.ThumbnailService

	TrackNormalizer   *logic.TrackNormalizer
	TrackMatcher      *logic.TrackMatcher
	TrackService      *logic.TrackService
	SourceFileService *logic.SourceFileService
	SourceService     *logic.SourceService
	TapeService       *logic.TapeService
	AutoImportService *logic.AutoImportService

	SearchService    *logic.SearchService
	SongCacheService *logic.SongCacheService

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
	if context.ExternalPlaylistStorage, err = storage.NewExternalPlaylistStorage(db); err != nil {
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
	context.TrackMatcher = logic.NewTrackMatcher()
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
	context.AutoImportService = logic.NewAutoImportService(
		context.SourceService,
		context.TrackService,
		context.TrackMatcher,
	)

	context.SearchService = logic.NewSearchService(context.SourceStorage, context.TrackStorage)

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
	}

	context.SongCacheService = logic.NewSongCacheService(
		context.SubsonicProviders,
		context.CachedMuxSongStorage,
		context.TrackMatcher,
	)

	subsonicMux := logic.NewSubsonicMuxService(
		context.MuxedSongListensStorage,
		context.SongCacheService,
		util.TakeIf(context.ScrobbleService, config.ScrobbleMode == configPkg.ScrobbleAll),
	)
	context.SubsonicMuxer = subsonicMux

	for _, subsonicProvider := range context.SubsonicProviders {
		subsonicMux.AddService(subsonicProvider)
	}

	context.SubsonicService = logic.NewSubsonicMainService(
		subsonicMux,
		context.SubsonicProviders,
		context.CachedMuxSongStorage,
		context.CachedMuxAlbumStorage,
		context.CachedMuxArtistStorage,
		context.ExternalPlaylistStorage,
	)

	if err = registerBackgroundTasks(&context); err != nil {
		return nil, err
	}

	return &context, nil
}

func registerBackgroundTasks(context *Context) error {
	cron := cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
		cron.WithSeconds(),
	)

	type backgroundTaskAndConfig struct {
		task   tasks.BackgroundTask
		config configPkg.BackgroundTaskConfig
	}

	scheduledTasks := []backgroundTaskAndConfig{}

	scheduledTasks = append(
		scheduledTasks,
		backgroundTaskAndConfig{
			task: tasks.NewDownloadSourcesTaskHandler(
				context.SourceFileService,
				context.SourceStorage,
			),
			config: context.Config.TasksDownloadSources,
		},
	)

	if context.ListenBrainzClient != nil {
		scheduledTasks = append(
			scheduledTasks,
			backgroundTaskAndConfig{
				task: tasks.NewListenBrainzPlaylistSyncHandler(
					context.ListenBrainzClient,
					context.SongCacheService,
					context.ExternalPlaylistStorage,
				),
				config: context.Config.TasksListenBrainzPlaylistSync,
			},
		)
	} else {
		slog.Info("Not registering ListenBrainz playlist sync task because ListenBrainz client is not configured")
	}

	if context.LastFmClient != nil {
		scheduledTasks = append(
			scheduledTasks,
			backgroundTaskAndConfig{
				task: tasks.NewLastFmPlaylistSyncHandler(
					context.LastFmClient,
					context.LastFmService,
					context.SongCacheService,
					context.AutoImportService,
					context.ExternalPlaylistStorage,
					context.Config.LastFmTargetPlaylistSize,
				),
				config: context.Config.TasksLastFmPlaylistSync,
			},
		)
	} else {
		slog.Info("Not registering last.fm playlist sync task because last.fm client is not configured")
	}

	scheduledTasks = append(
		scheduledTasks,
		backgroundTaskAndConfig{
			task: tasks.NewSyncLibraryHandler(
				context.SubsonicProviders,
				context.CachedMuxSongStorage,
				context.CachedMuxAlbumStorage,
				context.CachedMuxArtistStorage,
			),
			config: context.Config.TasksSyncLibrary,
		},
	)

	for _, scheduledTask := range scheduledTasks {
		err := setupBackgroundTask(cron, scheduledTask.task, scheduledTask.config)
		if err != nil {
			return err
		}
	}

	cron.Start()

	return nil
}

func setupBackgroundTask(cron *cron.Cron, task tasks.BackgroundTask, config configPkg.BackgroundTaskConfig) error {
	if config.Cron == configPkg.CronDisabled {
		slog.Info(fmt.Sprintf("Background task %s is disabled, skipping cron registration", task.Name()))
		return nil
	}

	if config.MaxAttempts < 1 {
		slog.Warn(fmt.Sprintf("Max attempts for background task %s is set to %d < 1, forcing to 1", task.Name(), config.MaxAttempts))
		config.MaxAttempts = 1
	}

	_, err := cron.AddFunc(config.Cron, func() {
		for retries := 0; retries < config.MaxAttempts; retries++ {
			slog.Log(context.Background(), configPkg.LevelTrace, fmt.Sprintf("Running background task %s, attempt %d/%d", task.Name(), retries+1, config.MaxAttempts))

			err := task.OnSchedule()
			if err == nil {
				slog.Log(context.Background(), configPkg.LevelTrace, fmt.Sprintf("Background task %s succeeded on attempt %d/%d", task.Name(), retries+1, config.MaxAttempts))
				return
			}

			if retries == config.MaxAttempts-1 {
				slog.Error(fmt.Sprintf("Background task %s failed after %d retries, giving up: %s", task.Name(), config.MaxAttempts, err.Error()))
				return
			}

			slog.Warn(fmt.Sprintf("Attempt %d/%d for background task %s failed, will retry in %.0fs: %s", retries+1, config.MaxAttempts, task.Name(), config.RetryDelay.Seconds(), err.Error()))
			time.Sleep(config.RetryDelay)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to setup background task %s: %w", task.Name(), err)
	}

	slog.Info(fmt.Sprintf("Registered background task %s with cron=%s", task.Name(), config.Cron))
	return nil
}
