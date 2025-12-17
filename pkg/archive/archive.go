// Package archive provides utilities for reading archives
package archive

import (
	"archive/zip"
	"os"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

type Archive struct {
	Path    string
	Size    int64
	ModTime time.Time
	Type    string
	Files   []File
}

type File struct {
	Path    string
	ModTime time.Time
	Mime    string
	Size    int64
}

func ArchiveFromZip(path string) (Archive, error) {
	finf, err := os.Lstat(path)
	if err != nil {
		return Archive{}, err
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		return Archive{}, err
	}

	arc := Archive{
		Path:    path,
		Size:    finf.Size(),
		ModTime: finf.ModTime(),
		Type:    "zip",
	}

	arc.Files = make([]File, 0, len(zr.File))
	for _, zf := range zr.File {
		zfInf := zf.FileInfo()
		if zfInf.IsDir() {
			continue
		}
		zfr, err := zf.Open()
		if err != nil {
			return Archive{}, err
		}

		m, err := mimetype.DetectReader(zfr)
		if err != nil {
			zfr.Close()
			return Archive{}, err
		}

		arc.Files = append(arc.Files, File{
			Path:    zf.Name,
			ModTime: zfInf.ModTime(),
			Size:    zfInf.Size(),
			Mime:    m.String(),
		})
		zfr.Close()
	}

	return arc, nil
}
