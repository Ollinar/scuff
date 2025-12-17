package vips_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/Ollinar/scuff/internal/image"
	vips_processor "github.com/Ollinar/scuff/internal/vips"
)

func TestVipProcessor_Save(t *testing.T) {
	vpp := vips_processor.NewVipsImageProcessor(slog.Default())
	err := vpp.Start(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer vpp.Shutdown()

	fSrc, err := os.Open("./../../test/images/02.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer fSrc.Close()

	outF, err := os.Create("./../../test/images/02out.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer outF.Close()
	err = vpp.Start(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer vpp.Shutdown()
	imgR, err := vpp.Load(t.Context(), fSrc)
	if err != nil {
		t.Fatal(err)
	}
	defer imgR.Close()
	//
	// dim, err := vpp.Dimension(t.Context(), imgR)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("dimension: %+v", dim)

	resImg, err := vpp.Resize(t.Context(), imgR, 500, 1000, true)
	if err != nil {
		t.Fatal(err)
	}
	defer resImg.Close()

	err = vpp.Save(t.Context(), resImg, outF, image.GIF)
	if err != nil {
		t.Fatal(err)
	}
}
