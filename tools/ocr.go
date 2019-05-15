package tools

import (
	"bufio"
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
	txt string
	orientation int

	txtErr         error
	orientationErr error
)

func processText(wg *sync.WaitGroup,path string,language string,imageFile string) {
	defer wg.Done()

	outputFile,err := common.CreateTempFile()
	if err != nil {
		txtErr = err
		return
	}
	defer common.FileDelete(outputFile.Name())

	cmd := exec.Command(filepath.Join(path,"tesseract.exe"),"--tessdata-dir",filepath.Join(path,"tessdata"),imageFile,outputFile.Name(),"-l",language)
	if common.IsDebugMode() {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		txtErr = err
		return
	}

	b,err := common.FileExists(outputFile.Name() + ".txt")
	if err != nil {
		txtErr = err
		return
	}
	if !b {
		txtErr = &common.ErrFileNotFound{outputFile.Name() + ".txt"}
	}

	if cmd.ProcessState.ExitCode() != 0 {
		txtErr = &errors.ErrProcessState{cmd.ProcessState.ExitCode()}
		return
	}

	buf,err := ioutil.ReadFile(outputFile.Name() + ".txt")
	if err != nil {
		txtErr = err
		return
	}

	txt = string(buf)
}

func processOrientation(wg *sync.WaitGroup,path string,language string,imageFile string) {
	defer wg.Done()

	outputFile,err := common.CreateTempFile()
	if err != nil {
		orientationErr = err
		return
	}
	defer common.FileDelete(outputFile.Name())

	cmd := exec.Command(filepath.Join(path,"tesseract"),"--tessdata-dir",filepath.Join(path,"tessdata"),imageFile,outputFile.Name(),"-l",language,"--psm","0")
	if common.IsDebugMode() {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Run()
	if err != nil {
		orientationErr = err
		return
	}

	if cmd.ProcessState.ExitCode() != 0 {
		orientationErr = &errors.ErrProcessState{cmd.ProcessState.ExitCode()}
		return
	}

	b,err := common.FileExists(outputFile.Name() + ".osd")
	if err != nil {
		txtErr = err
		return
	}
	if !b {
		txtErr = &common.ErrFileNotFound{outputFile.Name() + ".osd"}
	}

	buf,err := ioutil.ReadFile(outputFile.Name() + ".osd")
	if err != nil {
		orientationErr = err
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
			orientation,err = strconv.Atoi(line)
			break
		}
	}
}

func Ocr(path string,language string,imageFile string) (string,int,error) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go processText(&wg,path,language,imageFile)

	wg.Add(1)
	go processOrientation(&wg,path,language,imageFile)

	wg.Wait()

	if txtErr != nil {
		return "",-1, txtErr
	}

	if orientationErr != nil {
		return "",-1,orientationErr
	}

	return txt,orientation,nil
}