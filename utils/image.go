package utils

import (
	"flag"
	"fmt"
	"golang.org/x/image/tiff"
	"image"
	"image/draw"
	"io"
	"os"

	"github.com/mpetavy/common"

	"github.com/disintegration/imaging"
)

var quality *int

type Orientation int

const (
	ORIENTATION_0 Orientation = iota
	ORIENTATION_90
	ORIENTATION_180
	ORIENTATION_270
)

type Rotation int

const (
	ROTATE_0 Rotation = iota
	ROTATE_90
	ROTATE_180
	ROTATE_270
)

func init() {
	quality = flag.Int("jpeg.quality", 80, "JPEG quality")
}

func LoadImage(path string) (img image.Image, err error) {
	defer func() {
		if err := recover(); err != nil {
			img = nil
			err = fmt.Errorf("LoadImage failed: %s", path)
		}
	}()

	img, err = imaging.Open(path)

	if err == nil {
		return img, nil
	}

	f, err := os.Open(path)
	if common.Error(err) {
		return nil, err
	}

	defer func() {
		common.Ignore(f.Close())
	}()

	img, err = tiff.Decode(f)
	if common.Error(err) {
		return nil, err
	}

	return img, nil
}

func CopyImage(src image.Image) draw.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	return dst
}

func Rotate(source image.Image, rotation Rotation) image.Image {
	switch rotation {
	case ROTATE_0:
		return CopyImage(source)
	case ROTATE_90:
		return imaging.Rotate90(source)
	case ROTATE_180:
		return imaging.Rotate180(source)
	case ROTATE_270:
		return imaging.Rotate270(source)
	}

	return source
}

func EncodeJpeg(source image.Image, w io.Writer) error {
	return imaging.Encode(w, source, imaging.JPEG, imaging.JPEGQuality(*quality))
}

func SaveJpeg(source image.Image, filename string) error {
	f, err := os.Create(filename)
	if common.Error(err) {
		return err
	}

	defer func() {
		common.Ignore(f.Close())
	}()

	return EncodeJpeg(source, f)
}
