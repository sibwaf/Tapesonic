package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"tapesonic/util"
)

type Ffmpeg struct {
	path string
}

func NewFfmpeg(path string) *Ffmpeg {
	return &Ffmpeg{
		path: path,
	}
}

func (f *Ffmpeg) Stream(ctx context.Context, offsetMs int, durationMs int, reader *ReaderWithMeta) (io.ReadCloser, error) {
	cmd := exec.CommandContext(
		ctx,
		f.path,

		"-v", "0",

		"-ss", fmt.Sprintf("%.3f", float32(offsetMs)/1000.0),
		"-i", "-",
		"-t", fmt.Sprintf("%.3f", float32(durationMs)/1000.0),
		"-vn",
		"-f", "opus", // todo
		// todo: -c copy

		"-",
	)
	cmd.Stdin = reader.Reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start streaming `%s`: %s", reader.SourceInfo, err.Error()))
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start streaming `%s`: %s", reader.SourceInfo, err.Error()))
		return nil, err
	}

	slog.Debug(fmt.Sprintf("Streaming `%s` via `%s`", reader.SourceInfo, cmd.String()))

	return util.NewCustomCloseReadCloser(stdout, func() error {
		err = cmd.Wait()

		var exitError *exec.ExitError
		if errors.As(err, &exitError) && ctx.Err() == context.Canceled {
			slog.Debug(fmt.Sprintf("Stopped streaming `%s` because context was cancelled: %s", reader.SourceInfo, err.Error()))
			return nil
		} else if err != nil {
			slog.Error(fmt.Sprintf("Error while streaming `%s`: %s", reader.SourceInfo, err.Error()))
			if exitError != nil && len(exitError.Stderr) > 0 {
				slog.Error(string(exitError.Stderr))
			}
			return err
		} else {
			slog.Debug(fmt.Sprintf("Successfully finished streaming `%s`", reader.SourceInfo))
			return nil
		}
	}), nil
}
