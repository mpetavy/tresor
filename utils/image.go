package utils

import (
	"bytes"
	"flag"
	"fmt"
	chaiTiff "github.com/chai2010/tiff"
	defaultTiff "golang.org/x/image/tiff"
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

func LoadImage(ba []byte) (img image.Image, err error) {
	defer func() {
		if err := recover(); err != nil {
			img = nil
			err = fmt.Errorf("LoadImage failed")
		}
	}()

	img, err = imaging.Decode(bytes.NewReader(ba))

	if err == nil {
		return img, nil
	}

	img, err = defaultTiff.Decode(bytes.NewReader(ba))
	if common.Error(err) {
		return nil, err
	}

	img, err = chaiTiff.Decode(bytes.NewReader(ba))
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
		common.DebugError(f.Close())
	}()

	return EncodeJpeg(source, f)
}
