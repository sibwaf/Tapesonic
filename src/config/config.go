package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"tapesonic/util"
	"time"
)

type TapesonicConfig struct {
	ServerPort int
	Username   string
	Password   string

	WebappDir       string
	DataStorageDir  string
	MediaStorageDir string
	CacheDir        string

	YtdlpPath  string
	FfmpegPath string

	TasksImportQueueImport BackgroundTaskConfig
	TasksLibrarySync       BackgroundTaskConfig

	SubsonicProxyUrl      string
	SubsonicProxyUsername string
	SubsonicProxyPassword string

	StreamCacheSize        int64
	StreamCacheMinLifetime time.Duration
}

type BackgroundTaskConfig struct {
	Cron     string
	Cooldown time.Duration
}

func NewConfig() (*TapesonicConfig, error) {
	portText := getEnvOrDefault("TAPESONIC_PORT", "8080")
	port, err := strconv.Atoi(portText)
	if err != nil {
		return nil, fmt.Errorf("TAPESONIC_PORT is not a number: %s", portText)
	}

	config := &TapesonicConfig{
		ServerPort: port,
		Username:   os.Getenv("TAPESONIC_USERNAME"),
		Password:   os.Getenv("TAPESONIC_PASSWORD"),

		YtdlpPath:  getEnvOrDefault("TAPESONIC_YTDLP_PATH", "yt-dlp"),
		FfmpegPath: getEnvOrDefault("TAPESONIC_FFMPEG_PATH", "ffmpeg"),

		WebappDir:       getEnvOrDefault("TAPESONIC_WEBAPP_DIR", "webapp"),
		DataStorageDir:  getEnvOrDefault("TAPESONIC_DATA_STORAGE_DIR", "data"),
		MediaStorageDir: getEnvOrDefault("TAPESONIC_MEDIA_STORAGE_DIR", "media"),
		CacheDir:        getEnvOrDefault("TAPESONIC_CACHE_DIR", "cache"),

		TasksImportQueueImport: getBackgroundTaskConfig("IMPORT_QUEUE_IMPORT", "0 * * * * *", 15*time.Minute),
		TasksLibrarySync:       getBackgroundTaskConfig("LIBRARY_SYNC", "0 */15 * * * *", 15*time.Minute),

		SubsonicProxyUrl:      os.Getenv("TAPESONIC_SUBSONIC_PROXY_URL"),
		SubsonicProxyUsername: os.Getenv("TAPESONIC_SUBSONIC_PROXY_USERNAME"),
		SubsonicProxyPassword: os.Getenv("TAPESONIC_SUBSONIC_PROXY_PASSWORD"),

		StreamCacheSize:        getEnvSizeOrDefault("TAPESONIC_STREAM_CACHE_SIZE", 512*1024*1024), // 512 MB
		StreamCacheMinLifetime: getEnvDurationOrDefault("TAPESONIC_STREAM_CACHE_MIN_LIFETIME", 1*time.Hour),
	}

	return config, nil
}

func getBackgroundTaskConfig(
	name string,
	defaultCron string,
	defaultCooldown time.Duration,
) BackgroundTaskConfig {
	return BackgroundTaskConfig{
		Cron:     getEnvOrDefault(fmt.Sprintf("TAPESONIC_TASKS_%s_CRON", name), defaultCron),
		Cooldown: getEnvDurationOrDefault(fmt.Sprintf("TAPESONIC_TASKS_%s_COOLDOWN", name), defaultCooldown),
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
