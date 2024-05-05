package util

import "io"

type customReadCloser struct {
	reader io.Reader
	close  func() error
}

func NewCustomReadCloser(reader io.Reader, close func() error) io.ReadCloser {
	return &customReadCloser{
		reader: reader,
		close:  close,
	}
}

func (rc *customReadCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *customReadCloser) Close() error {
	return rc.close()
}
