package tools

import (
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"testing"
)

func TestMain(m *testing.M) {
	defer common.Cleanup()
	flag.Parse()
	common.Exit(m.Run())
}

func TestOcr(t *testing.T) {
	txt,orientation,err := Ocr("c:\\Tesseract-OCR","deu","d:\\archive\\sample\\15\\page.1")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(txt)
	fmt.Printf("Orientaton: %d\n",orientation)
}