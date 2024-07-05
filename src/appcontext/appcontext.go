package appcontext

import (
	"log/slog"
	"os"
	"path"
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/http/subsonic/client"
	"tapesonic/logic"
	"tapesonic/storage"
	"tapesonic/tasks"
	"tapesonic/ytdlp"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Context struct {
	Config *config.TapesonicConfig

	TapeStorage             *storage.TapeStorage
	TrackStorage            *storage.TrackStorage
	PlaylistStorage         *storage.PlaylistStorage
	AlbumStorage            *storage.AlbumStorage
	ImportQueueStorage      *storage.ImportQueueStorage
	TapeTrackListensStorage *storage.TapeTrackListensStorage
	CachedMuxSongStorage    *storage.CachedMuxSongStorage
	CachedMuxAlbumStorage   *storage.CachedMuxAlbumStorage
	CachedMuxArtistStorage  *storage.CachedMuxArtistStorage
	MuxedSongListensStorage *storage.MuxedSongListensStorage
	MediaStorage            *storage.MediaStorage
	Importer                *storage.Importer

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg

	SubsonicProviders []*logic.SubsonicNamedService
	SubsonicMuxer     logic.SubsonicService
	SubsonicService   logic.SubsonicService
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

	context.MediaStorage = storage.NewMediaStorage(
		config.MediaStorageDir,
		context.TapeStorage,
		context.PlaylistStorage,
		context.AlbumStorage,
	)

	context.Importer = storage.NewImporter(
		context.Config.MediaStorageDir,
		context.Ytdlp,
		context.TapeStorage,
	)

	subsonicMux := logic.NewSubsonicMuxService(
		context.CachedMuxSongStorage,
		context.MuxedSongListensStorage,
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
