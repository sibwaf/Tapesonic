package util

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

const (
	requestTraceHeader = "X-Attr-RequestTraceId"
)

var sensitiveFields = []string{
	SUBSONIC_QUERY_PASSWORD,
	SUBSONIC_QUERY_TOKEN,
	SUBSONIC_QUERY_SALT,
}

func Logged(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()
		r.Header.Add(requestTraceHeader, requestId)

		LogInfo(r, fmt.Sprintf("Serving request %s %s", r.Method, maskUrl(*r.URL)))
		handler(w, r)
	}
}

func maskUrl(u url.URL) string {
	params := u.Query()
	for _, name := range sensitiveFields {
		if params.Has(name) {
			params.Set(name, "xxxxxx")
		}
	}

	result := url.URL(u)
	result.RawQuery = params.Encode()
	return result.String()
}

func LogDebug(r *http.Request, message string, args ...any) {
	slog.Debug(message, prepareArgs(r, args)...)
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
