// Package image defines the seperation between the app and the image processor it uses
package image

import (
	"context"
	"io"
)

type ImageType int

const (
	JPG ImageType = 1 + iota
	PNG
	WEBP
	JXL
	GIF
)

const SOURCE ImageType = 0

type Dimension struct {
	Width  int
	Height int
}

type ImageRef interface {
	// Close is used to cleanup any resources the ImageRef uses
	Close() error
}

type Processor interface {
	IsSupported(mime string) bool
	Dimension(ctx context.Context, img ImageRef) (Dimension, error)

	// TODO: make these methods more configurable
	Resize(ctx context.Context, img ImageRef, width int, height int, keepAspectRatio bool) (ImageRef, error)
	Load(ctx context.Context, src io.ReadCloser) (ImageRef, error)
	Save(ctx context.Context, img ImageRef, dest io.Writer, imgT ImageType) error
	Start(ctx context.Context) error
	Shutdown() error
}
