package ytdlp

import (
	"encoding/json"
	"log/slog"
	"os/exec"
)

type Ytdlp struct {
	path string
}

func NewYtdlp(path string) *Ytdlp {
	return &Ytdlp{
		path: path,
	}
}

func (y *Ytdlp) ExtractMetadata(url string) (YtdlpMetadata, error) {
	slog.Debug("Extracting metadata", "yt-dlp", y.path, "url", url)

	cmd := exec.Command(y.path, "--dump-json", url)

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug("Failed to extract metadata", "yt-dlp", y.path, "url", url, "error", err)
		return YtdlpMetadata{}, err
	}

	slog.Debug("Successfully extracted metadata", "yt-dlp", y.path, "url", url)

	var metadata YtdlpMetadata
	err = json.Unmarshal(out, &metadata)
	return metadata, err
}

func (y *Ytdlp) Download(url string, formatId string, downloadDir string) (YtdlpMetadata, error) {
	slog.Debug("Downloading files", "yt-dlp", y.path, "url", url, "path", downloadDir)

	args := []string{
		"--format", formatId,
		"--convert-thumbnails", "png",
		"--output", "%(id)s.%(ext)s",
		"--paths", "home:" + downloadDir,

		"--no-continue",
		"--no-part",
		"--no-simulate",
		"--write-thumbnail",
		"--write-info-json",
		"--dump-json",

		url,
	}

	cmd := exec.Command(y.path, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug("Failed to download files", "yt-dlp", y.path, "url", url, "path", downloadDir, "error", err)
		return YtdlpMetadata{}, err
	}

	slog.Debug("Successfully downloaded files", "yt-dlp", y.path, "url", url, "path", downloadDir)

	var metadata YtdlpMetadata
	err = json.Unmarshal(out, &metadata)
	return metadata, err
}
