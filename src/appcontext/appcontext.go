package appcontext

import (
	"os"
	"path"

	configPkg "tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/lastfm"
	"tapesonic/library"
	"tapesonic/listenbrainz"
	"tapesonic/logic"
	"tapesonic/media"
	"tapesonic/recommendations"
	"tapesonic/remotes"
	"tapesonic/scheduling"
	"tapesonic/scrobbling"
	"tapesonic/search"
	"tapesonic/sources"
	"tapesonic/storage"
	"tapesonic/tapes"
	"tapesonic/users"
	"tapesonic/ytdlp"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Context struct {
	Config     *configPkg.TapesonicConfig
	Scheduling *scheduling.SchedulingModule

	Ytdlp  *ytdlp.Ytdlp
	Ffmpeg *ffmpeg.Ffmpeg

	Users           *users.UsersModule
	Sources         *sources.SourcesModule
	Remotes         *remotes.RemotesModule
	Tapes           *tapes.TapesModule
	ListenBrainz    *listenbrainz.ListenBrainzModule
	LastFm          *lastfm.LastFmModule
	Library         *library.LibraryModule
	Media           *media.MediaModule
	Scrobbling      *scrobbling.ScrobblingModule
	Search          *search.SearchModule
	Recommendations *recommendations.RecommendationsModule

	SourceStorage     *storage.SourceStorage
	SourceFileStorage *storage.SourceFileStorage
	TrackStorage      *storage.TrackStorage
	ThumbnailStorage  *storage.ThumbnailStorage

	YtdlpMetadataStorage *storage.YtdlpMetadataStorage
	MediaStorage         *storage.MediaStorage
	StreamCacheStorage   *storage.StreamCacheStorage

	YtdlpService *logic.YtdlpService

	ThumbnailService *logic.ThumbnailService

	TrackNormalizer   *logic.TrackNormalizer
	TrackMatcher      *logic.TrackMatcher
	TrackService      *logic.TrackService
	SourceFileService *logic.SourceFileService
	SourceService     *logic.SourceService
	AutoImportService *logic.AutoImportService
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

	if err := db.Exec(`DROP VIEW IF EXISTS all_playlist_tracks`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`DROP VIEW IF EXISTS all_playlists`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`DROP VIEW IF EXISTS all_albums`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`DROP VIEW IF EXISTS all_tracks`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`DROP VIEW IF EXISTS all_artists`).Error; err != nil {
		return nil, err
	}
	if err := db.Exec(`DROP VIEW IF EXISTS all_covers`).Error; err != nil {
		return nil, err
	}

	context.Scheduling = scheduling.NewSchedulingModule(db, config.SchedulerDelay)

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
	if context.YtdlpMetadataStorage, err = storage.NewYtdlpMetadataStorage(db); err != nil {
		return nil, err
	}

	if err = storage.Migrate(db); err != nil {
		return nil, err
	}

	context.YtdlpService = logic.NewYtdlpService(
		context.Ytdlp,
		context.YtdlpMetadataStorage,
		config.YtdlpMetadataMaxLifetime,
		config.YtdlpMetadataMaxParallelism,
	)

	context.SourceFileService = logic.NewSourceFileService(
		context.SourceFileStorage,
		context.SourceStorage,
		context.YtdlpService,
		config.MediaStorageDir,
	)

	context.ThumbnailService = logic.NewThumbnailService(
		context.ThumbnailStorage,
		path.Join(config.MediaStorageDir, "thumbnails"),
	)

	if context.Users, err = users.NewUsersModule(db); err != nil {
		return nil, err
	}
	context.Sources = sources.NewSourcesModule(context.SourceFileService, context.SourceStorage, config.DownloadNextSourceSchedule)
	if context.Remotes, err = remotes.NewRemotesModule(
		db,
		context.Scheduling.TaskScheduler,
		config.RemoteLibrarySyncSchedule,
	); err != nil {
		return nil, err
	}
	context.Tapes = tapes.NewTapesModule(db, context.TrackStorage)
	context.Library = library.NewLibraryModule(db)

	context.MediaStorage = storage.NewMediaStorage(db, config.MediaStorageDir)

	if context.StreamCacheStorage, err = storage.NewStreamCacheStorage(
		path.Join(config.CacheDir, "stream"),
		config.StreamCacheSize,
		config.StreamCacheMinLifetime,
		db,
	); err != nil {
		return nil, err
	}

	context.ListenBrainz = listenbrainz.NewListenBrainzModule(db)

	if context.LastFm, err = lastfm.NewLastFmModule(db, config.LastFmApiKey, config.LastFmApiSecret); err != nil {
		return nil, err
	}

	context.TrackNormalizer = logic.NewTrackNormalizer()
	context.TrackMatcher = logic.NewTrackMatcher()
	context.TrackService = logic.NewTrackService(context.TrackStorage)

	context.SourceService = logic.NewSourceService(
		context.SourceStorage,
		context.YtdlpService,
		context.SourceFileService,
		context.TrackService,
		context.ThumbnailService,
		context.TrackNormalizer,
	)
	context.AutoImportService = logic.NewAutoImportService(
		context.SourceService,
		context.TrackService,
		context.TrackMatcher,
	)

	context.Scrobbling = scrobbling.NewScrobblingModule(
		db,
		context.ListenBrainz.ListenBrainzService,
		context.LastFm.LastFmService,
		context.Library.LibraryService,
		context.Remotes.RemoteService,
	)

	context.Media = media.NewMediaModule(
		context.Remotes.RemoteService,
		context.ThumbnailService,
		context.MediaStorage,
		context.StreamCacheStorage,
		context.Ffmpeg,
		context.YtdlpService,
	)

	context.Search = search.NewSearchModule(
		db,
		context.Library.LibraryService,
		context.SourceService,
		context.TrackService,
		context.AutoImportService,
		context.TrackMatcher,
	)

	context.Recommendations = recommendations.NewRecommendationModule(
		db,
		context.LastFm.LastFmRecommendationService,
		context.ListenBrainz.ListenBrainzRecommendationService,
		context.Search.SearchService,
		context.Scheduling.TaskScheduler,
		config.RecommendationPlaylistSyncSchedule,
	)

	if err = prepareDatabase(&context); err != nil {
		return nil, err
	}

	registerSchedulers(&context)
	// todo: stop schedulers

	return &context, nil
}

func prepareDatabase(context *Context) error {
	if err := context.Scheduling.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Tapes.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Scrobbling.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Recommendations.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Library.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.ListenBrainz.PrepareDatabase(); err != nil {
		return err
	}
	return nil
}

func registerSchedulers(context *Context) {
	context.Sources.RegisterSchedules(context.Scheduling.TaskScheduler)
	context.Remotes.RegisterSchedules(context.Scheduling.TaskScheduler)
	context.Recommendations.RegisterSchedules(context.Scheduling.TaskScheduler)
}
