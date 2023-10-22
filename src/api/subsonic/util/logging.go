package util

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const (
	requestTraceHeader = "X-Attr-RequestTraceId"
)

func Logged(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()
		r.Header.Add(requestTraceHeader, requestId)

		LogInfo(r, "Serving request", "method", r.Method, "path", r.URL.Path, "client", GetClientName(r))
		handler(w, r)
	}
}

func LogInfo(r *http.Request, message string, args ...any) {
	slog.Info(message, prepareArgs(r, args)...)
}

func LogWarning(r *http.Request, message string, args ...any) {
	slog.Warn(message, prepareArgs(r, args)...)
}

func LogError(r *http.Request, message string, args ...any) {
	slog.Error(message, prepareArgs(r, args)...)
}

func prepareArgs(r *http.Request, args []any) []any {
	fullArgs := make([]any, 0, len(args)+2)
	fullArgs = append(fullArgs, "id", r.Header.Get(requestTraceHeader))
	fullArgs = append(fullArgs, args...)
	return fullArgs
}
