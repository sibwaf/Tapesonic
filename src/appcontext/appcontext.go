package appcontext

import (
	"os"
	"path"
	configPkg "tapesonic/config"
	"tapesonic/ffmpeg"
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
	TrackStorage                *storage.TrackStorage
	PlaylistStorage             *storage.PlaylistStorage
	AlbumStorage                *storage.AlbumStorage
	ImportQueueStorage          *storage.ImportQueueStorage
	TapeTrackListensStorage     *storage.TapeTrackListensStorage
	CachedMuxSongStorage        *storage.CachedMuxSongStorage
	CachedMuxAlbumStorage       *storage.CachedMuxAlbumStorage
	CachedMuxArtistStorage      *storage.CachedMuxArtistStorage
	MuxedSongListensStorage     *storage.MuxedSongListensStorage
	ListenbrainzPlaylistStorage *storage.ListenbrainzPlaylistStorage
	MediaStorage                *storage.MediaStorage
	StreamCacheStorage          *storage.StreamCacheStorage
	Importer                    *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg

	ListenBrainzClient *listenbrainz.ListenBrainzClient

	SubsonicProviders []*logic.SubsonicNamedService
	SubsonicMuxer     logic.SubsonicService
	SubsonicService   logic.SubsonicService

	StreamService   *logic.StreamService
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
		sqlite.Open(path.Join(config.DataStorageDir, "data.sqlite?_foreign_keys=on")),
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

	context.TapeStorage, err = storage.NewTapeStorage(db)
	if err != nil {
		return nil, err
	}
	if context.TrackStorage, err = storage.NewTrackStorage(db); err != nil {
		return nil, err
	}
	context.PlaylistStorage, err = storage.NewPlaylistStorage(db)
	if err != nil {
		return nil, err
	}
	if context.AlbumStorage, err = storage.NewAlbumStorage(db); err != nil {
		return nil, err
	}
	if context.ImportQueueStorage, err = storage.NewImportQueueStorage(db); err != nil {
		return nil, err
	}
	if context.TapeTrackListensStorage, err = storage.NewTapeTrackListensStorage(db); err != nil {
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

	context.MediaStorage = storage.NewMediaStorage(
		config.MediaStorageDir,
		context.TapeStorage,
		context.PlaylistStorage,
		context.AlbumStorage,
	)

	if context.StreamCacheStorage, err = storage.NewStreamCacheStorage(path.Join(config.CacheDir, "stream"), db); err != nil {
		return nil, err
	}

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
		context.TapeStorage,
	)

	if config.ListenBrainzToken != "" {
		context.ListenBrainzClient = listenbrainz.NewListenBrainzClient(config.ListenBrainzToken)
	}

	if context.ListenBrainzClient != nil {
		context.ScrobbleService = logic.NewScrobbleService(*context.ListenBrainzClient)
	}

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
			context.TapeTrackListensStorage,
			context.MediaStorage,
			context.Ffmpeg,
			util.TakeIf(context.ScrobbleService, config.ScrobbleMode == configPkg.ScrobbleTapesonic),
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

	context.StreamService = logic.NewStreamService(
		context.SubsonicService,
		context.StreamCacheStorage,
		config.StreamCacheSize,
		config.StreamCacheMinLifetime,
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

	if err = tasks.NewImportQueueTaskHandler(
		context.ImportQueueStorage,
		context.Importer,
		context.Config.TasksImportQueueImport,
	).RegisterSchedules(cron); err != nil {
		return err
	}

	if context.ListenBrainzClient != nil {
		if err = tasks.NewListenBrainzPlaylistSyncHandler(
			context.ListenBrainzClient,
			context.CachedMuxSongStorage,
			*context.ListenbrainzPlaylistStorage,
			context.Config.TasksListenBrainzPlaylistSync,
		).RegisterSchedules(cron); err != nil {
			return err
		}
	}

	if err = tasks.NewSyncLibraryHandler(
		context.SubsonicProviders,
		context.CachedMuxSongStorage,
		context.CachedMuxAlbumStorage,
		context.CachedMuxArtistStorage,
		context.Config.TasksLibrarySync,
	).RegisterSchedules(cron); err != nil {
		return err
	}

	cron.Start()

	return nil
}
