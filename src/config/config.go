package config

import (
	"fmt"
	"os"
	"strconv"
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
