package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func (f *Ffmpeg) Stream(ctx context.Context, offsetMs int, durationMs int, reader io.Reader) (io.ReadCloser, error) {
	ctx, cancel := context.WithCancel(ctx)

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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start streaming via `%s`: %s", cmd.String(), err.Error())
	}

	err = cmd.Start()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start streaming via `%s`: %s", cmd.String(), err.Error())
	}

	return &ffmpegReader{cancel: cancel, cmd: cmd, stdout: stdout}, nil
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
		if errors.As(err, &exitError) {
			errorMsg := fmt.Sprintf("error while streaming via `%s`: %s", reader.cmd.String(), err.Error())
			if len(exitError.Stderr) > 0 {
				err = fmt.Errorf("%s (%s)", errorMsg, string(exitError.Stderr))
			} else {
				err = errors.New(errorMsg)
			}
		} else {
			err = fmt.Errorf("error while streaming via `%s`: %s", reader.cmd.String(), err.Error())
		}

		return n, err
	}

	return n, err
}

func (reader *ffmpegReader) Close() error {
	reader.cancel()
	return reader.cmd.Wait()
}
