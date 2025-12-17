package app_test

import (
	"io"
	"testing"

	"github.com/Ollinar/scuff/internal/app"
	"github.com/Ollinar/scuff/internal/image"
)

func Test_pageModule_GenerateThumbnail(t *testing.T) {
	ap, err := app.NewApp(t.Context(), "./../../test/test.db", nil,
		app.WithContentDirectory("./../../test/content/"))
	if err != nil {
		t.Fatal(err)
	}
	defer ap.Close()

	pgs, err := ap.Page().GetAll(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(pgs) < 10 {
		return
	}

	pg := pgs[4]

	_ = pg

	err = ap.Page().GeneratePage(t.Context(), noopWriter{Writer: io.Discard}, pg, 500, 1000, image.JPG)
	if err != nil {
		t.Fatal(err)
	}
}

type noopWriter struct {
	io.Writer
}

func (npw noopWriter) Close() error {
	return nil
}
