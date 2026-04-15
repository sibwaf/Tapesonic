package appcontext

import (
	"os"
	"path"

	"tapesonic/artists"
	"tapesonic/artworks"
	configPkg "tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/lastfm"
	"tapesonic/library"
	"tapesonic/listenbrainz"
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

	Ytdlp  *ytdlp.YtdlpModule
	Ffmpeg *ffmpeg.Ffmpeg

	Users           *users.UsersModule
	Artworks        *artworks.ArtworksModule
	Artists         *artists.ArtistsModule
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

	StreamCacheStorage *storage.StreamCacheStorage
}

func NewContext(config *configPkg.TapesonicConfig) (*Context, error) {
	var err error
	context := Context{
		Config: config,
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

	if err := storage.DropCascades(db); err != nil {
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
	if err := db.Exec(`DROP VIEW IF EXISTS all_artworks`).Error; err != nil {
		return nil, err
	}

	context.Scheduling = scheduling.NewSchedulingModule(db, config.SchedulerDelay)

	if err = storage.Migrate(db); err != nil {
		return nil, err
	}

	context.Ytdlp = ytdlp.NewYtdlpModule(
		db,
		config.YtdlpPath,
		config.YtdlpMetadataMaxLifetime,
		config.YtdlpMetadataMaxParallelism,
	)

	if context.Users, err = users.NewUsersModule(db); err != nil {
		return nil, err
	}
	context.Artworks = artworks.NewArtworksModule(
		db,
		path.Join(config.MediaStorageDir, "artworks"),
	)
	context.Artists = artists.NewArtistsModule(db)
	context.Sources = sources.NewSourcesModule(
		db,
		context.Ytdlp.YtdlpService,
		context.Artworks.ArtworkService,
		config.DownloadNextSourceSchedule,
		config.MediaStorageDir,
	)
	if context.Remotes, err = remotes.NewRemotesModule(
		db,
		context.Scheduling.TaskScheduler,
		context.Artists.ArtistService,
		config.RemoteLibrarySyncSchedule,
	); err != nil {
		return nil, err
	}
	context.Library = library.NewLibraryModule(db)
	context.Tapes = tapes.NewTapesModule(db, context.Artists.ArtistService, context.Sources.SourceService, context.Library.LibraryService)

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

	context.Scrobbling = scrobbling.NewScrobblingModule(
		db,
		context.ListenBrainz.ListenBrainzService,
		context.LastFm.LastFmService,
		context.Library.LibraryService,
		context.Remotes.RemoteService,
	)

	context.Media = media.NewMediaModule(
		context.Remotes.RemoteService,
		context.Artworks.ArtworkService,
		context.Sources.SourceService,
		context.StreamCacheStorage,
		context.Ffmpeg,
		context.Ytdlp.YtdlpService,
	)

	context.Search = search.NewSearchModule(
		db,
		context.Artists.ArtistService,
		context.Library.LibraryService,
		context.Sources.SourceService,
		context.Sources.ImportService,
	)

	context.Recommendations = recommendations.NewRecommendationModule(
		db,
		context.LastFm.LastFmRecommendationService,
		context.ListenBrainz.ListenBrainzRecommendationService,
		context.Search.SearchService,
		context.Scheduling.TaskScheduler,
		config.RecommendationPlaylistSyncSchedule,
	)

	if err = prepareDatabase(&context, db); err != nil {
		return nil, err
	}

	registerSchedulers(&context)
	// todo: stop schedulers

	return &context, nil
}

func prepareDatabase(context *Context, db *gorm.DB) error {
	if err := context.Scheduling.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Ytdlp.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Artworks.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Artists.PrepareDatabase(); err != nil {
		return err
	}
	if err := context.Sources.PrepareDatabase(); err != nil {
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

	if err := storage.CreateCascades(db); err != nil {
		return err
	}

	return nil
}

func registerSchedulers(context *Context) {
	context.Sources.RegisterSchedules(context.Scheduling.TaskScheduler)
	context.Remotes.RegisterSchedules(context.Scheduling.TaskScheduler)
	context.Recommendations.RegisterSchedules(context.Scheduling.TaskScheduler)
}
