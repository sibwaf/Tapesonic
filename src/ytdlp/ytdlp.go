package ytdlp

import (
	"encoding/json"
	"fmt"
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
	cmd := exec.Command(y.path, "--dump-json", url)

	slog.Debug(fmt.Sprintf("Extracting metadata from `%s` via `%s`", url, cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to extract metadata from `%s`: %s", url, err.Error()))
		outText := string(out)
		if outText != "" {
			slog.Error(outText)
		}
		return YtdlpMetadata{}, err
	}

	slog.Debug(fmt.Sprintf("Successfully extracted metadata from `%s`", url))

	var metadata YtdlpMetadata
	err = json.Unmarshal(out, &metadata)
	return metadata, err
}

func (y *Ytdlp) Download(url string, formatId string, downloadDir string) (YtdlpMetadata, error) {
	cmd := exec.Command(
		y.path,

		"--format", formatId,
		"--convert-thumbnails", "png",
		"--output", "%(id)s.%(ext)s",
		"--paths", "home:"+downloadDir,

		"--no-continue",
		"--no-part",
		"--no-simulate",
		"--write-thumbnail",
		"--write-info-json",
		"--dump-json",

		url,
	)

	slog.Debug(fmt.Sprintf("Downloading `%s` via `%s`", url, cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to download `%s`: %s", url, err.Error()))
		outText := string(out)
		if outText != "" {
			slog.Error(outText)
		}
		return YtdlpMetadata{}, err
	}

	slog.Debug(fmt.Sprintf("Successfully downloaded `%s`", url))

	var metadata YtdlpMetadata
	err = json.Unmarshal(out, &metadata)
	return metadata, err
}
