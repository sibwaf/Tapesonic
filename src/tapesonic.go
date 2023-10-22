package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"tapesonic/api"
	"tapesonic/build"
	"tapesonic/config"
)

var logo = []string{
	" ______                            _      ",
	"/_  __/__ ____  ___ ___ ___  ___  (_)___  ",
	" / / / _ `/ _ \\/ -_|_-</ _ \\/ _ \\/ / __/  ",
	"/_/  \\_,_/ .__/\\__/___/\\___/_//_/_/\\__/   ",
	"        /_/                               ",
}

func main() {
	for _, line := range logo {
		println(line)
	}

	slog.Info("Starting Tapesonic", "version", build.TAPESONIC_VERSION)

	port := os.Getenv("TAPESONIC_PORT")
	if port == "" {
		port = "8080"
	}

	config := &config.TapesonicConfig{
		Username: os.Getenv("TAPESONIC_USERNAME"),
		Password: os.Getenv("TAPESONIC_PASSWORD"),
	}

	mux := http.NewServeMux()
	for route, handler := range api.GetHandlers(config) {
		mux.HandleFunc(route, handler)
	}

	slog.Info("Serving HTTP requests", "port", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
	if err != nil && err != http.ErrServerClosed {
		slog.Error("Failed to serve requests", "error", err)
		os.Exit(1)
	}
}
