package ffmpeg

import "io"

type ReaderWithMeta struct {
	SourceInfo string
	Reader     io.Reader
}

func NewReaderWithMeta(
	sourceInfo string,
	reader io.Reader,
) *ReaderWithMeta {
	return &ReaderWithMeta{
		SourceInfo: sourceInfo,
		Reader:     reader,
	}
}
