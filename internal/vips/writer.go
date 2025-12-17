package vips

import (
	"errors"
	"io"
)

// wraps a writer with noop closer
type noopWriterCloser struct {
	err error
	io.Writer
}

func (nwc *noopWriterCloser) Write(p []byte) (int, error) {
	n, err := nwc.Writer.Write(p)
	if err != nil {
		nwc.err = errors.Join(err, nwc.err)
		return n, err
	}
	return n, nil
}

func (nwc noopWriterCloser) Close() error {
	return nil
}
