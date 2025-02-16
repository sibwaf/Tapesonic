package util

import (
	"fmt"
	"log/slog"
)

func FormatToMediaType(format string) string {
	switch format {
	case "flac":
		return "audio/flac"
	case "mp3":
		return "audio/mpeg"
	case "opus":
		return "audio/opus"
	case "m4a":
		return "audio/mp4"
	case "matroska":
		return "audio/x-matroska"
	case "webm":
		return "audio/x-matroska"

	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	}

	slog.Warn(fmt.Sprintf("No media type mapping for format `%s`", format))
	return "application/octet-stream"
}

func MediaTypeToFormat(mediaType string) string {
	switch mediaType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpeg"
	case "image/webp":
		return "webp"
	}

	slog.Warn(fmt.Sprintf("Unknown MIME type `%s`", mediaType))
	return "bin"
}
