package app_test

import (
	"crypto"
	"testing"

	"github.com/Ollinar/scuff/internal/app"
)

func Test_archiveModule_ScanContentDirectory(t *testing.T) {
	ap, err := app.NewApp(t.Context(), "./../../test/test.db", nil,
		app.WithHasher(crypto.SHA256),
		app.WithPartialHashLength(app.MiB),
		app.WithContentDirectory("./../../test/content/"),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = ap.Archive().ScanContentDirectory(t.Context())
	if err != nil {
		t.Fatal(err)
	}
}
