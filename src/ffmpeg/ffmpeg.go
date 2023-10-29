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

// todo: traceable reader
func (f *Ffmpeg) Stream(ctx context.Context, offsetMs int, durationMs int, reader io.Reader, writer io.Writer) error {
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
	cmd.Stdin = reader
	cmd.Stdout = writer

	slog.Debug("Starting streaming", "ffmpeg", f.path)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start streaming: %s", err.Error()), "ffmpeg", f.path)
		return err
	}

	err = cmd.Wait()

	var exitError *exec.ExitError
	if errors.As(err, &exitError) && ctx.Err() == context.Canceled {
		slog.Debug("ffmpeg was killed because context was cancelled", "ffmpeg", f.path)
		return nil
	} else if err != nil {
		slog.Error(fmt.Sprintf("Error while streaming: %s", err.Error()), "ffmpeg", f.path)
		if exitError != nil && len(exitError.Stderr) > 0 {
			slog.Error(string(exitError.Stderr))
		}
		return err
	} else {
		slog.Debug("Successfully finished streaming", "ffmpeg", f.path)
		return nil
	}
}
