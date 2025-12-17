package vips

import (
	"github.com/Ollinar/scuff/vips"
)

type loaderFunc func(src *vips.Source, opts any) (*vips.Image, error)

var loaderMap = map[string]loaderFunc{
	"image/jpeg": func(src *vips.Source, opts any) (*vips.Image, error) {
		op, ok := opts.(*vips.JpegloadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewJpegloadSource(src, op)
	},
	"image/png": func(src *vips.Source, opts any) (*vips.Image, error) {
		op, ok := opts.(*vips.PngloadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewPngloadSource(src, op)
	},
	"image/jxl": func(src *vips.Source, opts any) (*vips.Image, error) {
		// JPEG-XL format
		op, ok := opts.(*vips.JxlloadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewJxlloadSource(src, op)
	},
	"image/gif": func(src *vips.Source, opts any) (*vips.Image, error) {
		// Assuming there is a GIF loading function in the vips library
		op, ok := opts.(*vips.GifloadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewGifloadSource(src, op)
	},
	"image/webp": func(src *vips.Source, opts any) (*vips.Image, error) {
		op, ok := opts.(*vips.WebploadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewWebploadSource(src, op)
	},
	"image/heif": func(src *vips.Source, opts any) (*vips.Image, error) {
		op, ok := opts.(*vips.HeifloadSourceOptions)
		if !ok {
			op = nil
		}
		return vips.NewHeifloadSource(src, op)
	},
}
