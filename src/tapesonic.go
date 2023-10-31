package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"tapesonic/appcontext"
	"tapesonic/build"
	"tapesonic/config"
	tshttp "tapesonic/http"
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
		WebappDir:       os.Getenv("TAPESONIC_WEBAPP_DIR"),
		MediaStorageDir: os.Getenv("TAPESONIC_MEDIA_STORAGE_DIR"),
	}
	if config.YtdlpPath == "" {
		config.YtdlpPath = "yt-dlp"
	}
	if config.FfmpegPath == "" {
		config.FfmpegPath = "ffmpeg"
	}
	if config.WebappDir == "" {
		config.WebappDir = "webapp"
	}
	if config.MediaStorageDir == "" {
		config.MediaStorageDir = "media"
	}

	appCtx := appcontext.NewContext(config)

	mux := http.NewServeMux()
	for route, handler := range tshttp.GetHandlers(appCtx) {
		mux.HandleFunc(route, handler)
	}

	slog.Info("Serving HTTP requests", "port", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to serve requests", "error", err)
		os.Exit(1)
	}
}
