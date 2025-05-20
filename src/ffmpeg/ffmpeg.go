package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"tapesonic/config"
)

var (
	// alac/mp4a are not supported by Chromium-based clients, don't support streaming them
	// mp4 format is not supported by ffmpeg stdout output; use "matroska" instead for alac/mp4

	codecToFormat = map[string]string{
		"flac": "flac",
		"mp3":  "mp3",
		"opus": "opus",
	}
	formatToCodecs = map[string][]string{
		"flac": {"flac"},
		"mp3":  {"mp3"},
		"opus": {"opus"},
	}

	versionRegexp = regexp.MustCompile(`ffmpeg version ([^\s]+)`)
)

const (
	ANY_FORMAT      = ""
	SEEKABLE_FORMAT = "mp3" // opus produces different binaries on each encode which breaks seeking
	FALLBACK_FORMAT = "opus"
)

type Ffmpeg struct {
	path string
}

func NewFfmpeg(path string) *Ffmpeg {
	return &Ffmpeg{
		path: path,
	}
}

func (f *Ffmpeg) GetCurrentVersion() (string, error) {
	cmd := exec.Command(f.path, "-version")

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	match := versionRegexp.FindSubmatch(out)
	if match == nil {
		return "", fmt.Errorf("can't extract version from '%s'", out)
	}

	return string(match[1]), nil
}

func (f *Ffmpeg) StreamFrom(
	ctx context.Context,
	sourceCodec string,
	targetFormat string,
	offsetMs int64,
	durationMs int64,
	input string,
) (format string, reader io.ReadCloser, err error) {
	// ignore codec parameters, only use codec name (ex. "mp4a.40.2")
	sourceCodec = strings.Split(sourceCodec, ".")[0]

	ctx, cancel := context.WithCancel(ctx)

	args := []string{}
	args = append(args, "-v", "0")

	if offsetMs > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", float32(offsetMs)/1000.0))
	}

	args = append(args, "-i", input)
	args = append(args, "-t", fmt.Sprintf("%.3f", float32(durationMs)/1000.0))
	args = append(args, "-vn")

	if targetFormat == ANY_FORMAT {
		targetFormat = codecToFormat[sourceCodec]
		if targetFormat == "" {
			targetFormat = FALLBACK_FORMAT
		}
	}

	if targetFormat == "opus" && offsetMs > 0 {
		// ffmpeg somehow fails to copy the audio data from
		// youtube-encoded opus if the starting position is not 0,
		// so we have to reencode it
		// https://stackoverflow.com/questions/60621646
		targetFormat = "opus"
	} else if slices.Contains(formatToCodecs[targetFormat], sourceCodec) {
		args = append(args, "-c:a", "copy")
	}

	args = append(args, "-f", targetFormat)
	args = append(args, "-")

	cmd := exec.CommandContext(ctx, f.path, args...)
	slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Streaming via ffmpeg: %s", cmd.String()))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return "", nil, fmt.Errorf("failed to start streaming via `%s`: %w", cmd.String(), err)
	}

	err = cmd.Start()
	if err != nil {
		cancel()
		return "", nil, fmt.Errorf("failed to start streaming via `%s`: %w", cmd.String(), err)
	}

	return targetFormat, &ffmpegReader{cancel: cancel, cmd: cmd, stdout: stdout}, nil
}

type ffmpegReader struct {
	cancel context.CancelFunc
	cmd    *exec.Cmd
	stdout io.Reader
}

func (reader *ffmpegReader) Read(p []byte) (n int, err error) {
	n, err = reader.stdout.Read(p)

	if err == io.EOF {
		err = reader.cmd.Wait()
		if err == nil {
			return n, io.EOF
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) && len(exitError.Stderr) > 0 {
			err = fmt.Errorf("error while streaming via `%s`: (%s) %w", reader.cmd.String(), string(exitError.Stderr), err)
		} else {
			err = fmt.Errorf("error while streaming via `%s`: %w", reader.cmd.String(), err)
		}

		return n, err
	}

	return n, err
}

func (reader *ffmpegReader) Close() error {
	reader.cancel()
	return reader.cmd.Wait()
}
