package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"tapesonic/api"
	"tapesonic/appcontext"
	"tapesonic/build"
	"tapesonic/config"
	"tapesonic/ffmpeg"
	"tapesonic/storage"
	"tapesonic/ytdlp"
)

var logo = []string{
	" ______                            _      ",
	"/_  __/__ ____  ___ ___ ___  ___  (_)___  ",
	" / / / _ `/ _ \\/ -_|_-</ _ \\/ _ \\/ / __/  ",
	"/_/  \\_,_/ .__/\\__/___/\\___/_//_/_/\\__/   ",
	"        /_/                               ",
}

func main() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(logHandler))

	for _, line := range logo {
		println(line)
	}

	slog.Info("Starting Tapesonic", "version", build.TAPESONIC_VERSION)

	port := os.Getenv("TAPESONIC_PORT")
	if port == "" {
		port = "8080"
	}

	config := &config.TapesonicConfig{
		Username:        os.Getenv("TAPESONIC_USERNAME"),
		Password:        os.Getenv("TAPESONIC_PASSWORD"),
		YtdlpPath:       os.Getenv("TAPESONIC_YTDLP_PATH"),
		FfmpegPath:      os.Getenv("TAPESONIC_FFMPEG_PATH"),
		MediaStorageDir: os.Getenv("TAPESONIC_MEDIA_STORAGE_DIR"),
	}
	if config.YtdlpPath == "" {
		config.YtdlpPath = "yt-dlp"
	}
	if config.FfmpegPath == "" {
		config.FfmpegPath = "ffmpeg"
	}
	if config.MediaStorageDir == "" {
		config.MediaStorageDir = "media"
	}

	appCtx := &appcontext.Context{
		Config:  config,
		Storage: storage.NewStorage(config.MediaStorageDir),
		Ytdlp:   ytdlp.NewYtdlp(config.YtdlpPath),
		Ffmpeg:  ffmpeg.NewFfmpeg(config.FfmpegPath),
	}

	mux := http.NewServeMux()
	for route, handler := range api.GetHandlers(appCtx) {
		mux.HandleFunc(route, handler)
	}

	slog.Info("Serving HTTP requests", "port", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to serve requests", "error", err)
		os.Exit(1)
	}
}
