package util

import (
	"fmt"
	"log/slog"
)

func FormatToMediaType(format string) string {
	switch format {
	case "opus":
		return "audio/opus"
	case "png":
		return "image/png"
	case "jpg":
		return "image/jpeg"
	}

	slog.Warn(fmt.Sprintf("No media type mapping for format `%s`", format))
	return "application/octet-stream"
}
