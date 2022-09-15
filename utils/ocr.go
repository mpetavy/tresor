package utils

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	ocrOrientationTimeout = flag.Int("ocr.orientation.timeout", 3000, "OCR orientation timeout")
	ocrScanTimeout = flag.Int("ocr.scan.timeout", 5000, "OCR scan timeout")
}

func ScanText(imageFile string) (string, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "-l", *tesseractLanguage)

	ba, err := common.NewWatchdogCmd(cmd, common.MillisecondToDuration(*ocrScanTimeout))
	if common.Error(err) {
		return "", err
	}

	return string(ba), nil
}

type WordInfo struct {
	Level    int
	PageNum  int
	BlockNum int
	ParNum   int
	LineNum  int
	WordNum  int
	Left     int
	Top      int
	Width    int
	Height   int
	Conf     int
	Text     string
}

func intValue(s string) int {
	v, err := strconv.Atoi(s)
	common.Panic(err)

	return v
}

func ScanWordInfos(imageFile string) ([]WordInfo, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "-", "-l", *tesseractLanguage, "tsv")

	ba, err := common.NewWatchdogCmd(cmd, common.MillisecondToDuration(*ocrScanTimeout))
	if common.Error(err) {
		return nil, err
	}

	l := []WordInfo{}
	ok := false

	scanner := bufio.NewScanner(bytes.NewReader(ba))
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Fields(line)

		if len(cols) < 12 {
			continue
		}

		if !ok {
			_, err = strconv.Atoi(cols[0])
			if err != nil {
				continue
			}

			ok = true
		}

		wi := WordInfo{
			Level:    intValue(cols[0]),
			PageNum:  intValue(cols[1]),
			BlockNum: intValue(cols[2]),
			ParNum:   intValue(cols[3]),
			LineNum:  intValue(cols[4]),
			WordNum:  intValue(cols[5]),
			Left:     intValue(cols[6]),
			Top:      intValue(cols[7]),
			Width:    intValue(cols[8]),
			Height:   intValue(cols[9]),
			Conf:     intValue(cols[10]),
			Text:     cols[11],
		}

		l = append(l, wi)
	}

	return l, nil
}

func ScanOrientation(imageFile string) (Orientation, error) {
	cmd := exec.Command(*tesseractPath, imageFile, "stdout", "--psm", "0")

	ba, err := common.NewWatchdogCmd(cmd, common.MillisecondToDuration(*ocrOrientationTimeout))
	if common.Error(err) {
		return ORIENTATION_0, fmt.Errorf("cannot find orientation")
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
	orientation, err := ScanOrientation(imageFile)
	if err != nil {
		return "", -1, nil
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

		defer func() {
			common.Error(common.FileDelete(tmpFile.Name()))
		}()

		err = SaveJpeg(tmpImage, tmpFile.Name())
		if common.Error(err) {
			return "", -1, err
		}

		imageFile = tmpFile.Name()
	}

	txt, err := ScanText(imageFile)
	if common.Error(err) {
		return "", -1, err
	}

	wis, err := ScanWordInfos(imageFile)
	if common.Error(err) {
		return "", -1, err
	}

	for _, wi := range wis {
		fmt.Printf("%+v\n", wi)
	}

	return txt, orientation, nil
}
