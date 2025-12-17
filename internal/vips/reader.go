package vips

import (
	"errors"
	"io"
)

var ErrRead = errors.New("failed to read")

// meant to wrap a reader and provide a way to differentiate read error from
// libvips and reader error
type noopReadCloser struct {
	reader io.Reader
	err    error
}

func (vr *noopReadCloser) Read(b []byte) (n int, err error) {
	n, err = vr.reader.Read(b)
	if err != nil {
		vr.err = errors.Join(err, vr.err)
		return
	}
	return
}

func (vr noopReadCloser) Close() error {
	return nil
}
