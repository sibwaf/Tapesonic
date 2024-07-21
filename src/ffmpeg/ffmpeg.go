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

var (
	codecToFormat = map[string]string{
		// "alac": "mp4", // todo: mp4/m4a don't support exporting to stdout
		"flac": "flac",
		"mp3":  "mp3",
		"opus": "opus",
	}
)

type Ffmpeg struct {
	path string
}

func NewFfmpeg(path string) *Ffmpeg {
	return &Ffmpeg{
		path: path,
	}
}

func (f *Ffmpeg) StreamFromFile(
	ctx context.Context,
	sourceCodec string,
	offsetMs int,
	durationMs int,
	sourcePath string,
) (mediaType string, reader io.ReadCloser, err error) {
	ctx, cancel := context.WithCancel(ctx)

	args := []string{}
	args = append(args, "-v", "0")
	args = append(args, "-ss", fmt.Sprintf("%.3f", float32(offsetMs)/1000.0))
	args = append(args, "-i", sourcePath)
	args = append(args, "-t", fmt.Sprintf("%.3f", float32(durationMs)/1000.0))
	args = append(args, "-vn")

	format := codecToFormat[sourceCodec]
	if format == "opus" && offsetMs > 0 {
		// ffmpeg somehow fails to copy the audio data from
		// youtube-encoded opus if the starting position is not 0,
		// so we have to reencode it
		// https://stackoverflow.com/questions/60621646
		format = "opus"
	} else if format != "" {
		args = append(args, "-c:a", "copy")
	} else {
		format = "opus"
	}
	args = append(args, "-f", format)
	args = append(args, "-")

	cmd := exec.CommandContext(ctx, f.path, args...)
	slog.Debug(fmt.Sprintf("Streaming a file via ffmpeg: %s", cmd.String()))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		err = fmt.Errorf("failed to start streaming via `%s`: %s", cmd.String(), err.Error())
		return
	}

	err = cmd.Start()
	if err != nil {
		cancel()
		err = fmt.Errorf("failed to start streaming via `%s`: %s", cmd.String(), err.Error())
		return
	}

	return util.FormatToMediaType(format), &ffmpegReader{cancel: cancel, cmd: cmd, stdout: stdout}, nil
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
