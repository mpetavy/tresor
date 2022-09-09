package utils

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mpetavy/common"
)

var (
	tesseractPath         *string
	tesseractLanguage     *string
	ocrOrientationTimeout *int
	ocrScanTimeout        *int
)

func init() {
	tpath, err := exec.LookPath("tesseract")
	common.Panic(err)

	tesseractPath = flag.String("tesseract.path", tpath, "Tesseract path")
	tesseractLanguage = flag.String("tesseract.language", "deu", "Tesseract language")
	ocrOrientationTimeout = flag.Int("ocr.orientation.timeout", 30000, "OCR orientation timeout")
	ocrScanTimeout = flag.Int("ocr.scan.timeout", 50000, "OCR scan timeout")
}

func processText(imageFile string) (string, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "-l", *tesseractLanguage)

	ba, err := common.NewWatchdogCmd(cmd, time.Millisecond*time.Duration(*ocrScanTimeout))
	if common.Error(err) {
		return "", err
	}

	return string(ba), nil
}

func processOrientation(imageFile string) (Orientation, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "--psm", "0")

	ba, err := common.NewWatchdogCmd(cmd, time.Millisecond*time.Duration(*ocrOrientationTimeout))
	if common.Error(err) {
		return ORIENTATION_0, err
	}

	tags := []string{"Orientation in degrees:", "Orientation:"}
	s := string(ba)

	var line string
	var o int

	r := bufio.NewReader(strings.NewReader(s))
	for {
		line, err = r.ReadString('\n')
		if err == io.EOF {
			break
		}

		for _, tag := range tags {
			p := strings.Index(line, tag)
			if p != -1 {
				line = strings.TrimSpace(line[p+len(tag):])

				o, err = strconv.Atoi(line)
				if common.Error(err) {
					return ORIENTATION_0, err
				}

				switch o {
				case 0:
					return ORIENTATION_0, nil
				case 90:
					return ORIENTATION_270, nil
				case 180:
					return ORIENTATION_180, nil
				case 270:
					return ORIENTATION_90, nil
				}

				return ORIENTATION_0, fmt.Errorf("unknown orientation")
			}
		}
	}

	return ORIENTATION_0, nil
}

func Ocr(imageFile string) (string, Orientation, error) {
	orientation, err := processOrientation(imageFile)

	if common.Error(err) {
		return "", -1, err
	}

	if orientation != 0 {
		ba, err := os.ReadFile(imageFile)
		if common.Error(err) {
			return "", -1, err
		}

		tmpImage, err := LoadImage(ba)
		if common.Error(err) {
			return "", -1, err
		}

		switch orientation {
		case ORIENTATION_90:
			tmpImage = Rotate(tmpImage, ROTATE_270)
		case ORIENTATION_180:
			tmpImage = Rotate(tmpImage, ROTATE_180)
		case ORIENTATION_270:
			tmpImage = Rotate(tmpImage, ROTATE_90)
		}

		tmpFile, err := common.CreateTempFile()
		if common.Error(err) {
			return "", -1, err
		}

		err = SaveJpeg(tmpImage, tmpFile.Name())
		if common.Error(err) {
			return "", -1, err
		}

		imageFile = tmpFile.Name()
	}

	txt, err := processText(imageFile)

	if common.Error(err) {
		return "", -1, err
	}

	return txt, orientation, nil
}
