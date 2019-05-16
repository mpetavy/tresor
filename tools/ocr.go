package tools

import (
	"bufio"
	"flag"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/utils/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	orientation_in_degrees = "Orientation in degrees:"
)

var (
	tesseractPath *string
	tesseractLanguage *string
)

func init() {
	tesseractPath = flag.String("tesseract_path","c:\\Tesseract-OCR","Tesseract path")
	tesseractLanguage = flag.String("tesseract_language","deu","Tesseract language")
}

func processText(wg *sync.WaitGroup,path string,language string,imageFile string,txt *string,err error) {
	defer wg.Done()

	var outputFile *os.File

	outputFile,err = common.CreateTempFile()
	if err != nil {
		return
	}
	defer common.FileDelete(outputFile.Name())

	cmd := exec.Command(filepath.Join(path,"tesseract.exe"),"--tessdata-dir",filepath.Join(path,"tessdata"),imageFile,outputFile.Name(),"-l",language)
	if common.IsDebugMode() {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		return
	}

	var b bool

	b,err = common.FileExists(outputFile.Name() + ".txt")
	if err != nil {
		return
	}
	if !b {
		err = &common.ErrFileNotFound{outputFile.Name() + ".txt"}
	}

	if cmd.ProcessState.ExitCode() != 0 {
		err = &errors.ErrProcessState{cmd.ProcessState.ExitCode()}
		return
	}

	var buf []byte

	buf,err = ioutil.ReadFile(outputFile.Name() + ".txt")
	if err != nil {
		return
	}

	*txt = string(buf)
}

func processOrientation(wg *sync.WaitGroup,path string,language string,imageFile string,orientation *int,err error) {
	defer wg.Done()

	var outputFile *os.File

	outputFile,err = common.CreateTempFile()
	if err != nil {
		return
	}
	defer common.FileDelete(outputFile.Name())

	cmd := exec.Command(filepath.Join(path,"tesseract"),"--tessdata-dir",filepath.Join(path,"tessdata"),imageFile,outputFile.Name(),"-l",language,"--psm","0")
	if common.IsDebugMode() {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		return
	}

	if cmd.ProcessState.ExitCode() != 0 {
		err = &errors.ErrProcessState{cmd.ProcessState.ExitCode()}
		return
	}

	var b bool

	b,err = common.FileExists(outputFile.Name() + ".osd")
	if err != nil {
		return
	}
	if !b {
		err = &common.ErrFileNotFound{outputFile.Name() + ".osd"}
		return
	}

	var buf []byte

	buf,err = ioutil.ReadFile(outputFile.Name() + ".osd")
	if err != nil {
		return
	}

	r := bufio.NewReader(strings.NewReader(string(buf)))
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		p := strings.Index(line,orientation_in_degrees)
		if p != -1 {
			line = strings.TrimSpace(line[p + len(orientation_in_degrees):])

			*orientation,err = strconv.Atoi(line)
			break
		}
	}
}

func Ocr(imageFile string) (string,int,error) {
	wg := sync.WaitGroup{}

	var txtErr error
	var txt string

	var orientationErr error
	var orientation int

	wg.Add(1)
	processText(&wg,*tesseractPath,*tesseractLanguage,imageFile,&txt,txtErr)

	wg.Add(1)
	processOrientation(&wg,*tesseractPath,*tesseractLanguage,imageFile,&orientation,orientationErr)

	wg.Wait()

	if txtErr != nil {
		return "",-1, txtErr
	}

	if orientationErr != nil {
		return "",-1,orientationErr
	}

	return txt,orientation,nil
}