// Package vips implement the image.Processor that the app uses
package vips

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	imageProc "github.com/Ollinar/scuff/internal/image"
	"github.com/Ollinar/scuff/vips"

	"github.com/gabriel-vasile/mimetype"
)

var debugMemoryStat = false

func EnableMemoryStat() {
	debugMemoryStat = true
}

func NewVipsImageProcessor(lggr *slog.Logger) VipProcessor {
	return VipProcessor{lggr: lggr}
}

type VipProcessor struct {
	lggr *slog.Logger
}

func (vpp VipProcessor) setLogger() {
	var level vips.LogLevel
	switch {
	case vpp.lggr.Enabled(context.Background(), slog.LevelError):
		level = vips.LogLevelError
	case vpp.lggr.Enabled(context.Background(), slog.LevelWarn):
		level = vips.LogLevelWarning
	case vpp.lggr.Enabled(context.Background(), slog.LevelInfo):
		level = vips.LogLevelInfo
	case vpp.lggr.Enabled(context.Background(), slog.LevelDebug):
		level = vips.LogLevelDebug
	}
	vips.SetLogging(func(messageDomain string, messageLevel vips.LogLevel, message string) {
		switch messageLevel {
		case vips.LogLevelDebug:
			vpp.lggr.Debug(message, slog.String("messageDomain", messageDomain))
		case vips.LogLevelCritical, vips.LogLevelError:
			vpp.lggr.Error(message, slog.String("messageDomain", messageDomain))
		case vips.LogLevelWarning:
			vpp.lggr.Warn(message, slog.String("messageDomain", messageDomain))
		case vips.LogLevelInfo:
			vpp.lggr.Info(message, slog.String("messageDomain", messageDomain))
		default:
		}
	}, level)

	if debugMemoryStat {
		tkr := time.Tick(time.Second)
		go func() {
			stat := &vips.MemoryStats{}
			for range tkr {
				vips.ReadVipsMemStats(stat)
				vpp.lggr.Debug("vips stat", slog.Any("stat", stat))
			}
		}()
	}
}

func (vpp VipProcessor) Start(ctx context.Context) error {
	vips.Startup(nil)
	vpp.setLogger()

	return nil
}

func (vpp VipProcessor) Shutdown() error {
	vips.Shutdown()
	return nil
}

func (vpp VipProcessor) Load(ctx context.Context, src io.ReadCloser) (imageProc.ImageRef, error) {
	imgR := &ImageRef{}
	headerBytes := make([]byte, 1024)

	n, err := src.Read(headerBytes)
	if err != nil && err != io.EOF {
		return nil, err
	}
	m := mimetype.Detect(headerBytes[:n])

	imgR.mime = m.String()
	imgR.headerBytes = headerBytes
	imgR.vipsR = &noopReadCloser{
		reader: io.MultiReader(bytes.NewReader(imgR.headerBytes), src),
	}

	imgR.src = vips.NewSource(imgR.vipsR)

	return imgR, nil
}

func (vpp VipProcessor) Resize(ctx context.Context, img imageProc.ImageRef, width int, height int, keepAspectRatio bool) (imageProc.ImageRef, error) {
	imgR, err := vpp.getVipImgRef(img)
	if err != nil {
		return nil, err
	}

	if imgR.img != nil {
		opts := vips.DefaultThumbnailImageOptions()
		opts.Height = height
		err := imgR.img.ThumbnailImage(width, opts)
		if err != nil {
			return nil, err
		}
	} else {
		opts := vips.DefaultThumbnailSourceOptions()
		if imgR.mime == "image/gif" || imgR.mime == "image/webp" {
			loadOp := &vips.LoadOptions{
				N: -1,
			}
			opts.OptionString = loadOp.OptionString()
		}
		opts.Height = height
		im, err := vips.NewThumbnailSource(imgR.src, width, opts)
		if err != nil {
			if imgR.vipsR.err != err {
				return nil, errors.Join(ErrRead, err)
			}
			return nil, err
		}
		imgR.img = im
	}

	return imgR, nil
}

func (vpp VipProcessor) IsSupported(mime string) bool {
	_, ok := loaderMap[mime]
	if !ok {
		return false
	}
	_, ok = saverMap[mime]
	return ok
}

func (vpp VipProcessor) Dimension(ctx context.Context, img imageProc.ImageRef) (imageProc.Dimension, error) {
	imgR, err := vpp.getVipImgRef(img)
	if err != nil {
		return imageProc.Dimension{}, err
	}
	if imgR.img == nil {
		err = imgR.loadImage()
		if err != nil {
			return imageProc.Dimension{}, err
		}
	}

	dim := imageProc.Dimension{
		Width:  imgR.img.Width(),
		Height: imgR.img.PageHeight(),
	}
	return dim, nil
}

func (vpp VipProcessor) Save(ctx context.Context, img imageProc.ImageRef, dest io.Writer, imgT imageProc.ImageType) error {
	imgR, err := vpp.getVipImgRef(img)
	if err != nil {
		return err
	}
	if imgR.img == nil {
		err = imgR.loadImage()
		if err != nil {
			return err
		}
	}
	var format string
	isAnimated := imgR.img.Pages() > 1
	var opts any = nil
	// NOTE: tweaking image option should be done here
	switch imgT {
	case imageProc.PNG:
		format = "image/png"
		if isAnimated {
			err = imgR.img.ExtractArea(0, 0, imgR.img.Width(), imgR.img.PageHeight())
			if err != nil {
				return err
			}
			err = imgR.img.SetPages(1)
			if err != nil {
				return err
			}
		}
	case imageProc.JPG:
		format = "image/jpeg"
		if isAnimated {
			err = imgR.img.ExtractArea(0, 0, imgR.img.Width(), imgR.img.PageHeight())
			if err != nil {
				return err
			}
			err = imgR.img.SetPages(1)
			if err != nil {
				return err
			}
		}
	case imageProc.JXL:
		format = "image/jxl"
		// NOTE: JXL does support animated images, but its wonky, so make it just save the first frame
		if isAnimated {
			err = imgR.img.ExtractArea(0, 0, imgR.img.Width(), imgR.img.PageHeight())
			if err != nil {
				return err
			}
			err = imgR.img.SetPages(1)
			if err != nil {
				return err
			}
		}
	case imageProc.GIF:
		format = "image/gif"
	case imageProc.WEBP:
		format = "image/webp"
	case imageProc.SOURCE:
		format = imgR.mime
	default:
		format = "image/jpeg"
		if isAnimated {
			err = imgR.img.ExtractArea(0, 0, imgR.img.Width(), imgR.img.PageHeight())
			if err != nil {
				return err
			}
			err = imgR.img.SetPages(1)
			if err != nil {
				return err
			}
		}
	}

	noopW := &noopWriterCloser{Writer: dest}

	t := vips.NewTarget(noopW)
	defer t.Close()

	fn, ok := saverMap[format]
	if !ok {
		return errors.New("image format is unsupported")
	}
	err = fn(imgR.img, t, opts)
	if err != nil {
		// check if its from writer
		if noopW.err != nil {
			return errors.Join(err, noopW.err)
		}
		return err
	}

	return nil
}

func (vpp VipProcessor) getVipImgRef(img imageProc.ImageRef) (*ImageRef, error) {
	if im, ok := img.(*ImageRef); ok {
		return im, nil
	}
	return nil, errors.New("img is not an image ref from vips processor")
}

type ImageRef struct {
	src         *vips.Source
	img         *vips.Image
	headerBytes []byte
	mime        string
	vipsR       *noopReadCloser
}

func (ir ImageRef) Close() error {
	ir.close()
	return nil
}

// loadImage will return an error when the src cant be loaded (format unsupported)
// if read of src fails it might also error
// vipsgen (maybe even libvips) doesn't provide destiction between them
func (ir *ImageRef) loadImage() error {
	loadFn, ok := loaderMap[ir.mime]
	if !ok {
		return errors.ErrUnsupported
	}
	var opts any
	// for those with multipages
	switch ir.mime {
	case "image/gif":
		opts = &vips.GifloadSourceOptions{
			N: -1,
		}
	case "image/webp":
		op := vips.DefaultWebploadSourceOptions()
		op.N = -1
		opts = op
	}
	img, err := loadFn(ir.src, opts)
	if err != nil {
		// check if the reader was the one to cause the error
		if ir.vipsR.err != nil {
			return errors.Join(ErrRead, err)
		}
		return err
	}
	ir.img = img
	return nil
}

func (ir *ImageRef) close() {
	if ir.img != nil {
		ir.img.Close()
		ir.img = nil
	}
	if ir.src != nil {
		ir.src.Close()
		ir.src = nil
	}
}
