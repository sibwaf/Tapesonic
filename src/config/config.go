package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"tapesonic/util"
	"time"
)

const (
	LevelTrace = slog.LevelDebug * 2

	ScrobbleNone      = 0
	ScrobbleTapesonic = 1
	ScrobbleAll       = 2

	CronDisabled = "off"
)

type TapesonicConfig struct {
	LogLevel slog.Level
	DevMode  bool

	ServerPort int
	Username   string
	Password   string

	WebappDir       string
	DataStorageDir  string
	MediaStorageDir string
	CacheDir        string

	YtdlpPath  string
	FfmpegPath string

	YtdlpMetadataMaxLifetime    time.Duration
	YtdlpMetadataMaxParallelism int

	TasksDownloadSources          BackgroundTaskConfig
	TasksSyncLibrary              BackgroundTaskConfig
	TasksListenBrainzPlaylistSync BackgroundTaskConfig
	TasksLastFmPlaylistSync       BackgroundTaskConfig

	ScrobbleMode int

	SubsonicProxyUrl      string
	SubsonicProxyUsername string
	SubsonicProxyPassword string

	StreamCacheSize        int64
	StreamCacheMinLifetime time.Duration

	ListenBrainzToken string

	LastFmApiKey             string
	LastFmApiSecret          string
	LastFmTargetPlaylistSize int
}

type BackgroundTaskConfig struct {
	Cron        string
	RetryDelay  time.Duration
	MaxAttempts int
}

func NewConfig() (*TapesonicConfig, error) {
	logLevel := slog.LevelInfo
	switch strings.ToUpper(getEnvOrDefault("TAPESONIC_LOG_LEVEL", "INFO")) {
	case "TRACE":
		logLevel = LevelTrace
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	}

	portText := getEnvOrDefault("TAPESONIC_PORT", "8080")
	port, err := strconv.Atoi(portText)
	if err != nil {
		return nil, fmt.Errorf("TAPESONIC_PORT is not a number: %s", portText)
	}

	scrobbleMode := ScrobbleNone
	switch strings.ToLower(getEnvOrDefault("TAPESONIC_SCROBBLE_MODE", "none")) {
	case "none":
		scrobbleMode = ScrobbleNone
	case "tapesonic":
		scrobbleMode = ScrobbleTapesonic
	case "all":
		scrobbleMode = ScrobbleAll
	}

	config := &TapesonicConfig{
		LogLevel: logLevel,
		DevMode:  getEnvBoolOrDefault("TAPESONIC_DEV_MODE", false),

		ServerPort: port,
		Username:   os.Getenv("TAPESONIC_USERNAME"),
		Password:   os.Getenv("TAPESONIC_PASSWORD"),

		YtdlpPath:  getEnvOrDefault("TAPESONIC_YTDLP_PATH", "yt-dlp"),
		FfmpegPath: getEnvOrDefault("TAPESONIC_FFMPEG_PATH", "ffmpeg"),

		YtdlpMetadataMaxLifetime:    getEnvDurationOrDefault("TAPESONIC_YTDLP_METADATA_MAX_LIFETIME", 15*time.Minute),
		YtdlpMetadataMaxParallelism: getEnvIntOrDefault("TAPESONIC_YTDLP_METADATA_MAX_PARALLELISM", 4),

		WebappDir:       getEnvOrDefault("TAPESONIC_WEBAPP_DIR", "webapp"),
		DataStorageDir:  getEnvOrDefault("TAPESONIC_DATA_STORAGE_DIR", "data"),
		MediaStorageDir: getEnvOrDefault("TAPESONIC_MEDIA_STORAGE_DIR", "media"),
		CacheDir:        getEnvOrDefault("TAPESONIC_CACHE_DIR", "cache"),

		TasksDownloadSources:          getBackgroundTaskConfig("DOWNLOAD_SOURCES", "0 * * * * *", 15*time.Minute, 1),
		TasksSyncLibrary:              getBackgroundTaskConfig("SYNC_LIBRARY", "0 */15 * * * *", 1*time.Minute, 5),
		TasksListenBrainzPlaylistSync: getBackgroundTaskConfig("LISTENBRAINZ_PLAYLIST_SYNC", "0 0 4 * * *", 15*time.Minute, 5),
		TasksLastFmPlaylistSync:       getBackgroundTaskConfig("LAST_FM_PLAYLIST_SYNC", "0 0 4 * * *", 15*time.Minute, 5),

		ScrobbleMode: scrobbleMode,

		SubsonicProxyUrl:      os.Getenv("TAPESONIC_SUBSONIC_PROXY_URL"),
		SubsonicProxyUsername: os.Getenv("TAPESONIC_SUBSONIC_PROXY_USERNAME"),
		SubsonicProxyPassword: os.Getenv("TAPESONIC_SUBSONIC_PROXY_PASSWORD"),

		StreamCacheSize:        getEnvSizeOrDefault("TAPESONIC_STREAM_CACHE_SIZE", 512*1024*1024), // 512 MB
		StreamCacheMinLifetime: getEnvDurationOrDefault("TAPESONIC_STREAM_CACHE_MIN_LIFETIME", 1*time.Hour),

		ListenBrainzToken: os.Getenv("TAPESONIC_LISTENBRAINZ_TOKEN"),

		LastFmApiKey:             os.Getenv("TAPESONIC_LAST_FM_API_KEY"),
		LastFmApiSecret:          os.Getenv("TAPESONIC_LAST_FM_API_SECRET"),
		LastFmTargetPlaylistSize: getEnvIntOrDefault("TAPESONIC_LAST_FM_TARGET_PLAYLIST_SIZE", 40),
	}

	return config, nil
}

func getBackgroundTaskConfig(
	name string,
	defaultCron string,
	defaultRetryDelay time.Duration,
	defaultMaxAttempts int,
) BackgroundTaskConfig {
	return BackgroundTaskConfig{
		Cron:        getEnvOrDefault(fmt.Sprintf("TAPESONIC_TASKS_%s_CRON", name), defaultCron),
		RetryDelay:  getEnvDurationOrDefault(fmt.Sprintf("TAPESONIC_TASKS_%s_RETRY_DELAY", name), defaultRetryDelay),
		MaxAttempts: getEnvIntOrDefault(fmt.Sprintf("TAPESONIC_TASKS_%s_MAX_ATTEMPTS", name), defaultMaxAttempts),
	}
}

func getEnvOrDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	} else {
		return defaultValue
	}
}

func getEnvBoolOrDefault(name string, defaultValue bool) bool {
	switch strings.ToLower(os.Getenv(name)) {
	case "true", "yes", "1":
		return true
	case "false", "no", "0":
		return false
	default:
		return defaultValue
	}
}

func getEnvIntOrDefault(name string, defaultValue int) int {
	value := os.Getenv(name)
	if value != "" {
		return util.StringToIntOrDefault(value, defaultValue)
	} else {
		return defaultValue
	}
}

func getEnvSizeOrDefault(name string, defaultValue int64) int64 {
	value := strings.ToLower(os.Getenv(name))

	multiplier := int64(1)
	switch {
	case strings.HasSuffix(value, "b"):
		multiplier = 1
		value = strings.TrimSuffix(value, "b")
	case strings.HasSuffix(value, "k"):
		multiplier = 1024
		value = strings.TrimSuffix(value, "k")
	case strings.HasSuffix(value, "m"):
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(value, "m")
	case strings.HasSuffix(value, "g"):
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(value, "g")
	}

	size := util.StringToInt64OrNull(value)
	if size == nil {
		return defaultValue
	}

	return (*size) * multiplier
}

func getEnvDurationOrDefault(name string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	if durationValue, err := time.ParseDuration(value); err != nil {
		return defaultValue
	} else {
		return durationValue
	}
}
