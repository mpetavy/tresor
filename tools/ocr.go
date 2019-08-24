package tools

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	tesseractPath         *string
	tesseractDataPath     *string
	tesseractLanguage     *string
	ocrOrientationTimeout *int
	ocrScanTimeout        *int
)

func init() {
	tesseractPath = flag.String("tesseract.path", "/bin/tesseract", "Tesseract path")
	tesseractDataPath = flag.String("tesseract.data.path", "", "Tesseract data path")
	tesseractLanguage = flag.String("tesseract.language", "deu", "Tesseract language")
	ocrOrientationTimeout = flag.Int("ocr.orientation.timeout", 3000, "OCR orientation timeout")
	ocrScanTimeout = flag.Int("ocr.scan.timeout", 5000, "OCR scan timeout")
}

func processText(imageFile string) (string, error) {
	if *tesseractDataPath == "" {
		f := filepath.Join(filepath.Dir(*tesseractPath), "tessdata")

		b, _ := common.FileExists(f)
		if b {
			*tesseractDataPath = f
		} else {
			return "", fmt.Errorf("tessetact data path not set")
		}
	}

	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "-l", *tesseractLanguage, "--tessdata-dir", *tesseractDataPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := common.Watchdog(cmd, time.Millisecond*time.Duration(*ocrScanTimeout))
	if err != nil {
		return "", err
	}

	return string(stdout.Bytes()), nil
}

func processOrientation(imageFile string) (common.Orientation, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "--psm", "0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := common.Watchdog(cmd, time.Millisecond*time.Duration(*ocrOrientationTimeout))
	if err != nil {
		return common.ORIENTATION_0, err
	}

	tags := []string{"Orientation in degrees:", "Orientation:"}
	s := string(stdout.Bytes())
	if s == "" {
		s = string(stderr.Bytes())
	}

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
				if err != nil {
					return common.ORIENTATION_0, err
				}

				switch o {
				case 0:
					return common.ORIENTATION_0, nil
				case 90:
					return common.ORIENTATION_270, nil
				case 180:
					return common.ORIENTATION_180, nil
				case 270:
					return common.ORIENTATION_90, nil
				}

				return common.ORIENTATION_0, fmt.Errorf("unknown orientation")
			}
		}
	}

	return common.ORIENTATION_0, nil
}

func Ocr(imageFile string) (string, common.Orientation, error) {
	orientation, err := processOrientation(imageFile)

	if err != nil {
		common.WarnError(err)
	}

	if orientation != 0 {
		tmpImage, err := common.LoadImage(imageFile)
		if err != nil {
			return "", -1, err
		}

		switch orientation {
		case common.ORIENTATION_90:
			tmpImage = common.Rotate(tmpImage, common.ROTATE_270)
		case common.ORIENTATION_180:
			tmpImage = common.Rotate(tmpImage, common.ROTATE_180)
		case common.ORIENTATION_270:
			tmpImage = common.Rotate(tmpImage, common.ROTATE_90)
		}

		tmpFile, err := common.CreateTempFile()
		if err != nil {
			return "", -1, err
		}

		err = common.SaveJpeg(tmpImage, tmpFile.Name())
		if err != nil {
			return "", -1, err
		}

		imageFile = tmpFile.Name()
	}

	txt, err := processText(imageFile)

	if err != nil {
		return "", -1, err
	}

	return txt, orientation, nil
}
