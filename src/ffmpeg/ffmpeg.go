package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

type Ffmpeg struct {
	path string
}

func NewFfmpeg(path string) *Ffmpeg {
	return &Ffmpeg{
		path: path,
	}
}

func (f *Ffmpeg) Stream(ctx context.Context, offsetMs int, durationMs int, reader *ReaderWithMeta, writer io.Writer) error {
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
	cmd.Stdout = writer

	slog.Debug(fmt.Sprintf("Starting streaming `%s` via `%s`", reader.SourceInfo, cmd.String()))

	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start streaming `%s`: %s", reader.SourceInfo, err.Error()))
		return err
	}

	err = cmd.Wait()

	var exitError *exec.ExitError
	if errors.As(err, &exitError) && ctx.Err() == context.Canceled {
		slog.Debug(fmt.Sprintf("Stopped streaming `%s` because context was cancelled", reader.SourceInfo))
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
}
