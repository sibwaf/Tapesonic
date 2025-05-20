package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strings"
	"tapesonic/config"
)

type Ytdlp struct {
	path string
}

func NewYtdlp(path string) *Ytdlp {
	return &Ytdlp{
		path: path,
	}
}

func (y *Ytdlp) GetCurrentVersion() (string, error) {
	cmd := exec.Command(y.path, "--version")

	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func (y *Ytdlp) ExtractMetadata(url string) (YtdlpFile, error) {
	cmd := exec.Command(
		y.path,

		"--dump-single-json",
		"--flat-playlist",
		"--yes-playlist",

		url,
	)

	slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Extracting metadata via ytdlp: %s", cmd.String()))

	out, err := cmd.Output()
	if err != nil {
		return YtdlpFile{}, err
	}

	result := YtdlpFile{}
	return result, json.Unmarshal(out, &result)
}

func (y *Ytdlp) Download(url string, format string, downloadDir string) (YtdlpFile, error) {
	cmd := exec.Command(
		y.path,

		"-f", format,

		"--dump-single-json",
		"--no-simulate",

		"--no-playlist",
		"--no-continue",
		"--no-part",

		"--output", path.Join(downloadDir, "%(extractor_key)s-%(id)s.%(ext)s"),

		// can't use already available metadata from cache due to a bug in yt-dlp
		// when in some cases (at least with Bandcamp) it determines extension as "unknown_video"
		url,
	)

	slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Downloading format=%s via ytdlp: %s", format, cmd.String()))

	out, err := cmd.Output()
	if err != nil {
		return YtdlpFile{}, err
	}

	result := YtdlpFile{}
	return result, json.Unmarshal(out, &result)
}

func (y *Ytdlp) GetFormatFromMetadata(metadata string, format string) (YtdlpFormat, error) {
	cmd := exec.Command(
		y.path,

		"-f", format,
		"--dump-single-json",
		"--load-info-json", "-",
	)

	slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Getting format=%s via ytdlp: %s", format, cmd.String()))

	cmd.Stdin = bytes.NewReader([]byte(metadata))

	out, err := cmd.Output()
	if err != nil {
		return YtdlpFormat{}, err
	}

	result := YtdlpFormat{}
	return result, json.Unmarshal(out, &result)
}
