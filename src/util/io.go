package util

import "io"

type customCloseReadCloser struct {
	io.Reader
	close func() error
}

type customCloseReadSeekCloser struct {
	io.ReadSeeker
	close func() error
}

func NewCustomCloseReadCloser(reader io.Reader, close func() error) io.ReadCloser {
	return &customCloseReadCloser{
		Reader: reader,
		close:  close,
	}
}

func (rc *customCloseReadCloser) Close() error {
	return rc.close()
}

func NewCustomCloseReadSeekCloser(reader io.ReadSeeker, close func() error) io.ReadSeekCloser {
	return &customCloseReadSeekCloser{
		ReadSeeker: reader,
		close:      close,
	}
}

func (rsc *customCloseReadSeekCloser) Close() error {
	return rsc.close()
}
