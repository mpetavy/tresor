package tools

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/mpetavy/common"
	"github.com/stretchr/testify/assert"
	"image"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const (
	hello_world = "Hello, this is OCR"
)

func TestMain(m *testing.M) {
	defer common.Cleanup()
	common.Test(m)
	common.Exit(m.Run())
}

func TestOcr(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1648, 2338))

	dc := gg.NewContextForImage(img)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("c:/windows/fonts/Arial.ttf", 96); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(hello_world, 1648/2, 200, 0.5, 0.5)
	dc.DrawStringAnchored(hello_world, 1648/2, 400, 0.5, 0.5)
	dc.DrawStringAnchored(hello_world, 1648/2, 600, 0.5, 0.5)

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(f.Name())

	img1 := dc.Image()

	for i := 0; i < 4; i++ {
		if i > 0 {
			img1 = common.Rotate(img1, common.ROTATE_90)
		}

		err := common.SaveJpeg(img1, f.Name())
		if err != nil {
			t.Fatal(err)
		}

		txt, orientation, err := Ocr(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.Index(txt, hello_world) != -1, "Must have recognized text "+hello_world)
		assert.True(t, i == int(orientation), "Must have recognized orientation")

		fmt.Println(txt)
		fmt.Printf("Orientation: %d\n", orientation)
	}
}
