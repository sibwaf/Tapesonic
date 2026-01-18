package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"tapesonic/util"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	LevelTrace   = slog.LevelDebug * 2
	CronDisabled = "off"
)

type TapesonicConfig struct {
	LogLevel slog.Level
	DevMode  bool

	ServerPort int

	WebappDir       string
	DataStorageDir  string
	MediaStorageDir string
	CacheDir        string

	SchedulerDelay time.Duration

	YtdlpPath  string
	FfmpegPath string

	YtdlpMetadataMaxLifetime    time.Duration
	YtdlpMetadataMaxParallelism int

	StreamCacheSize        int64
	StreamCacheMinLifetime time.Duration

	LastFmApiKey             string
	LastFmApiSecret          string
	LastFmTargetPlaylistSize int

	RemoteLibrarySyncSchedule          cron.Schedule
	RecommendationPlaylistSyncSchedule cron.Schedule
	DownloadNextSourceSchedule         cron.Schedule
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

	remoteLibrarySyncSchedule, err := parseCronSchedule(getEnvOrDefault("TAPESONIC_REMOTE_LIBRARY_SYNC_CRON", "0 0/15 * * * *"))
	if err != nil {
		return &TapesonicConfig{}, err
	}

	recommendationPlaylistSyncSchedule, err := parseCronSchedule(getEnvOrDefault("TAPESONIC_RECOMMENDATION_PLAYLIST_SYNC_CRON", "0 0 4 * * *"))
	if err != nil {
		return &TapesonicConfig{}, err
	}

	downloadSourceSchedule, err := parseCronSchedule(getEnvOrDefault("TAPESONIC_DOWNLOAD_NEXT_SOURCE_CRON", "0 * * * * *"))
	if err != nil {
		return &TapesonicConfig{}, err
	}

	config := &TapesonicConfig{
		LogLevel: logLevel,
		DevMode:  getEnvBoolOrDefault("TAPESONIC_DEV_MODE", false),

		ServerPort: port,

		WebappDir:       getEnvOrDefault("TAPESONIC_WEBAPP_DIR", "webapp"),
		DataStorageDir:  getEnvOrDefault("TAPESONIC_DATA_STORAGE_DIR", "data"),
		MediaStorageDir: getEnvOrDefault("TAPESONIC_MEDIA_STORAGE_DIR", "media"),
		CacheDir:        getEnvOrDefault("TAPESONIC_CACHE_DIR", "cache"),

		SchedulerDelay: getEnvDurationOrDefault("TAPESONIC_SCHEDULER_DELAY", 30*time.Second),

		YtdlpPath:  getEnvOrDefault("TAPESONIC_YTDLP_PATH", "yt-dlp"),
		FfmpegPath: getEnvOrDefault("TAPESONIC_FFMPEG_PATH", "ffmpeg"),

		YtdlpMetadataMaxLifetime:    getEnvDurationOrDefault("TAPESONIC_YTDLP_METADATA_MAX_LIFETIME", 15*time.Minute),
		YtdlpMetadataMaxParallelism: getEnvIntOrDefault("TAPESONIC_YTDLP_METADATA_MAX_PARALLELISM", 6),

		StreamCacheSize:        getEnvSizeOrDefault("TAPESONIC_STREAM_CACHE_SIZE", 512*1024*1024), // 512 MB
		StreamCacheMinLifetime: getEnvDurationOrDefault("TAPESONIC_STREAM_CACHE_MIN_LIFETIME", 1*time.Hour),

		LastFmApiKey:             os.Getenv("TAPESONIC_LASTFM_API_KEY"),
		LastFmApiSecret:          os.Getenv("TAPESONIC_LASTFM_API_SECRET"),
		LastFmTargetPlaylistSize: getEnvIntOrDefault("TAPESONIC_LASTFM_TARGET_PLAYLIST_SIZE", 40),

		RemoteLibrarySyncSchedule:          remoteLibrarySyncSchedule,
		RecommendationPlaylistSyncSchedule: recommendationPlaylistSyncSchedule,
		DownloadNextSourceSchedule:         downloadSourceSchedule,
	}

	return config, nil
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

func parseCronSchedule(value string) (cron.Schedule, error) {
	if value == CronDisabled {
		return nil, nil
	}

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	result, err := parser.Parse(value)
	if err != nil {
		return &cron.SpecSchedule{}, fmt.Errorf("invalid cron \"%s\": %w", value, err)
	} else {
		return result, nil
	}
}
