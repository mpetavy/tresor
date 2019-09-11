package utils

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/fogleman/gg"
	"github.com/mpetavy/common"
	"github.com/stretchr/testify/assert"
)

const (
	msg = "Hello, this is OCR by Tesseract!"
)

func TestMain(m *testing.M) {
	defer common.Done()
	common.Exit(m.Run())
}

func TestOcr(t *testing.T) {
	font := common.Eval(common.IsLinuxOS(), "/usr/share/fonts/TTF/DejaVuSans.ttf", "c:/windows/fonts/Arial.ttf").(string)

	img := image.NewRGBA(image.Rect(0, 0, 1648, 2338))

	dc := gg.NewContextForImage(img)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	if err := dc.LoadFontFace(font, 96); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(msg, 1648/2, 200, 0.5, 0.5)
	dc.DrawStringAnchored(msg, 1648/2, 400, 0.5, 0.5)
	dc.DrawStringAnchored(msg, 1648/2, 600, 0.5, 0.5)

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		common.IgnoreError(os.Remove(f.Name()))
	}()

	img1 := dc.Image()

	for i := 0; i < 4; i++ {
		if i > 0 {
			img1 = Rotate(img1, ROTATE_90)
		}

		err := SaveJpeg(img1, f.Name())
		if err != nil {
			t.Fatal(err)
		}

		txt, orientation, err := Ocr(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.Index(txt, msg) != -1, "Must have recognized text "+msg)
		assert.True(t, i == int(orientation), "Must have recognized orientation")

		fmt.Println(txt)
		fmt.Printf("Orientation: %d\n", orientation)
	}
}
