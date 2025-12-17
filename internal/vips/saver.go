package vips

import (
	"github.com/Ollinar/scuff/vips"
)

type saverFunc func(img *vips.Image, target *vips.Target, opts any) error

var saverMap = map[string]saverFunc{
	"image/jpeg": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.JpegsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.JpegsaveTarget(target, op)
	},
	"image/png": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.PngsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.PngsaveTarget(target, op)
	},
	"image/gif": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.GifsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.GifsaveTarget(target, op)
	},
	"image/jxl": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.JxlsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.JxlsaveTarget(target, op)
	},
	"image/webp": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.WebpsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.WebpsaveTarget(target, op)
	},
	"image/heif": func(img *vips.Image, target *vips.Target, opts any) error {
		op, ok := opts.(*vips.HeifsaveTargetOptions)
		if !ok {
			op = nil
		}
		return img.HeifsaveTarget(target, op)
	},
}
