package tools

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/mpetavy/common"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	tesseractPath     *string
	tesseractLanguage *string
)

func init() {
	tesseractPath = flag.String("tesseract.path", "c:\\Tesseract-OCR", "Tesseract path")
	tesseractLanguage = flag.String("tesseract.language", "deu", "Tesseract language")
}

func processText(wg *sync.WaitGroup, path string, language string, imageFile string, txt *string, err *error) {
	defer wg.Done()

	cmd := exec.Command(filepath.Join(path, "tesseract.exe"), imageFile, "stdout", "-l", language)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	*err = cmd.Run()
	if *err != nil {
		return
	}

	*txt = string(stdout.Bytes())
}

func processOrientation(wg *sync.WaitGroup, path string, language string, imageFile string, orientation *common.Orientation, err *error) {
	defer wg.Done()

	cmd := exec.Command(filepath.Join(path, "tesseract"), imageFile, "stdout", "--psm", "0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	*err = cmd.Run()
	if *err != nil {
		return
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
		line, *err = r.ReadString('\n')
		if *err == io.EOF {
			break
		}

		for _, tag := range tags {
			p := strings.Index(line, tag)
			if p != -1 {
				line = strings.TrimSpace(line[p+len(tag):])

				o, *err = strconv.Atoi(line)
				if *err != nil {
					return
				}

				switch o {
				case 0:
					*orientation = common.ORIENTATION_0
				case 90:
					*orientation = common.ORIENTATION_270
				case 180:
					*orientation = common.ORIENTATION_180
				case 270:
					*orientation = common.ORIENTATION_90
				}
				return
			}
		}
	}
}

func Ocr(imageFile string) (string, common.Orientation, error) {
	wg := sync.WaitGroup{}

	var txtErr error
	var txt string

	var orientationErr error
	var orientation common.Orientation

	wg.Add(1)
	processOrientation(&wg, *tesseractPath, *tesseractLanguage, imageFile, &orientation, &orientationErr)

	if orientationErr != nil {
		common.WarnError(orientationErr)
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

	wg.Add(1)
	processText(&wg, *tesseractPath, *tesseractLanguage, imageFile, &txt, &txtErr)

	wg.Wait()

	if txtErr != nil {
		return "", -1, txtErr
	}

	return txt, orientation, nil
}
