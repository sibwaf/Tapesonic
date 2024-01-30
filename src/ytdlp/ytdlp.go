package ytdlp

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"sort"

	"github.com/google/uuid"
)

type Ytdlp struct {
	path string
}

func NewYtdlp(path string) *Ytdlp {
	return &Ytdlp{
		path: path,
	}
}

func (y *Ytdlp) ExtractMetadata(url string) (YtdlpFile, error) {
	cmd := exec.Command(y.path, "--dump-json", url)

	slog.Debug(fmt.Sprintf("Extracting metadata from `%s` via `%s`", url, cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to extract metadata from `%s`: %s", url, err.Error()))
		outText := string(out)
		if outText != "" {
			slog.Error(outText)
		}
		return YtdlpFile{}, err
	}

	slog.Debug(fmt.Sprintf("Successfully extracted metadata from `%s`", url))

	return (&YtdlpFileWrapper{Raw: out}).Parse()
}

func (y *Ytdlp) Download(url string, formatId string, downloadDir string) (DownloadedPlaylist, error) {
	infoDir := path.Join(os.TempDir(), "tapesonic", uuid.New().String())
	defer os.RemoveAll(infoDir)

	cmd := exec.Command(
		y.path,

		"--format", formatId,
		"--convert-thumbnails", "png",

		"--output", path.Join(downloadDir, "%(extractor_key)s-%(id)s.%(ext)s"),
		"--output", "infojson:"+path.Join(infoDir, "%(id)s"),
		"--output", "pl_infojson:"+path.Join(infoDir, "playlist"),

		"--yes-playlist",

		"--no-continue",
		"--no-part",
		"--write-thumbnail",
		"--write-info-json",

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
		return DownloadedPlaylist{}, err
	}

	slog.Debug(fmt.Sprintf("Successfully downloaded `%s`", url))

	playlist, err := parseDownloadedPlaylist(path.Join(infoDir, "playlist.info.json"))
	if err != nil {
		return playlist, err
	}

	fileMetadataFiles, err := os.ReadDir(infoDir)
	if err != nil {
		return playlist, err
	}

	for _, fileMetadataFile := range fileMetadataFiles {
		if fileMetadataFile.Name() == "playlist.info.json" {
			continue
		}

		file, err := parseDownloadedFile(path.Join(infoDir, fileMetadataFile.Name()))
		if err != nil {
			return playlist, err
		}

		playlist.Files = append(playlist.Files, file)
	}

	sort.Slice(playlist.Files, func(i, j int) bool {
		leftMetadata, err := playlist.Files[i].Metadata.Parse()
		if err != nil {
			return false
		}
		rightMetadata, err := playlist.Files[j].Metadata.Parse()
		if err != nil {
			return false
		}

		return leftMetadata.PlaylistIndex < rightMetadata.PlaylistIndex
	})

	return playlist, nil
}

func parseDownloadedPlaylist(infoPath string) (DownloadedPlaylist, error) {
	var err error
	playlist := DownloadedPlaylist{}

	playlist.Metadata.Raw, err = os.ReadFile(infoPath)
	if err != nil && !os.IsNotExist(err) {
		return playlist, err
	}

	return playlist, nil
}

func parseDownloadedFile(infoPath string) (DownloadedFile, error) {
	var err error
	file := DownloadedFile{}

	file.Metadata.Raw, err = os.ReadFile(infoPath)
	if err != nil {
		return file, err
	}

	fileMetadata, err := file.Metadata.Parse()
	if err != nil {
		return file, err
	}

	file.MediaPath = fmt.Sprintf("%s-%s.%s", fileMetadata.ExtractorKey, fileMetadata.Id, fileMetadata.Ext)
	file.ThumbnailPath = fmt.Sprintf("%s-%s.%s", fileMetadata.ExtractorKey, fileMetadata.Id, "png")

	return file, nil
}
