package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type TapesonicConfig struct {
	ServerPort int
	Username   string
	Password   string

	WebappDir       string
	DataStorageDir  string
	MediaStorageDir string

	YtdlpPath  string
	FfmpegPath string

	TasksImportQueueImport BackgroundTaskConfig
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
		ServerPort:      port,
		Username:        os.Getenv("TAPESONIC_USERNAME"),
		Password:        os.Getenv("TAPESONIC_PASSWORD"),
		YtdlpPath:       getEnvOrDefault("TAPESONIC_YTDLP_PATH", "yt-dlp"),
		FfmpegPath:      getEnvOrDefault("TAPESONIC_FFMPEG_PATH", "ffmpeg"),
		WebappDir:       getEnvOrDefault("TAPESONIC_WEBAPP_DIR", "webapp"),
		DataStorageDir:  getEnvOrDefault("TAPESONIC_DATA_STORAGE_DIR", "data"),
		MediaStorageDir: getEnvOrDefault("TAPESONIC_MEDIA_STORAGE_DIR", "media"),

		TasksImportQueueImport: getBackgroundTaskConfig("IMPORT_QUEUE_IMPORT", "0 * * * * *", 15*time.Minute),
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
