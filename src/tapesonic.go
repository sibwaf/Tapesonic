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
	"/_  __/__  ___  ___ ___ ___  ___  (_)___  ",
	" / / / _ `/ _ \\/ -_|_-</ _ \\/ _ \\/ / __/  ",
	"/_/  \\_,_/ .__/\\__/___/\\___/_//_/_/\\__/   ",
	"        /_/                               ",
}

func main() {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Failed to parse config: %s\n", err.Error())
		os.Exit(3)
	}

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	})
	slog.SetDefault(slog.New(logHandler))

	for _, line := range logo {
		println(line)
	}

	slog.Info(fmt.Sprintf("Starting Tapesonic %s", build.TAPESONIC_VERSION))

	appCtx, err := appcontext.NewContext(config)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start the application context: %s", err.Error()))
		os.Exit(2)
	}

	mux := http.NewServeMux()
	for route, handler := range tshttp.GetHandlers(appCtx) {
		mux.HandleFunc(route, handler)
	}

	slog.Info(fmt.Sprintf("Serving HTTP requests @ port %d", config.ServerPort))
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), mux)
	if err != nil && err != http.ErrServerClosed {
		slog.Error(fmt.Sprintf("Failed to serve requests: %s", err.Error()))
		os.Exit(1)
	}
}
