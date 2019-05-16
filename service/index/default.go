package index

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/mpetavy/go-dicom"
	"github.com/mpetavy/go-dicom/dicomtag"
	"github.com/mpetavy/tresor/tools"
	"io/ioutil"
	"os"
)

const (
	DEFAULT_INDEXER = ""
)

type DefaultIndexer struct {
	name string
}

func NewDefaultIndexer() (*DefaultIndexer, error) {
	return &DefaultIndexer{}, nil
}

func (defaultIndexer *DefaultIndexer) Init(cfg *common.Jason) error {
	name, err := cfg.String("name")
	if err != nil {
		return err
	}
	defaultIndexer.name = name

	return nil
}

func (defaultIndexer *DefaultIndexer) Start() error {
	return nil
}

func (defaultIndexer *DefaultIndexer) Stop() error {
	return nil
}

func (defaultIndexer *DefaultIndexer) indexDicom(path string,buffer []byte,options *Options) (*Mapping, *[]byte,error) {
	mapping := make(Mapping)

	var imageFile *os.File
	var dataset *dicom.DataSet
	var err error

	if len(buffer) > 0 {
		dataset, err = dicom.ReadDataSetInBytes(buffer, dicom.ReadOptions{})
	} else {
		dataset, err = dicom.ReadDataSetFromFile(path, dicom.ReadOptions{})
	}

	if err == nil {
		representativeFrameNumber := 0

		elem,err := dataset.FindElementByTag(dicomtag.RepresentativeFrameNumber)
		if  err == nil {
			representativeFrameNumber = elem.Value[0].(int)
		}

		for _, elem := range dataset.Elements {
			if elem.Tag == dicomtag.PixelData {
				data := elem.Value[0].(dicom.PixelDataInfo)
				for i, frame := range data.Frames {
					if i == representativeFrameNumber {
						var err error

						imageFile,err = common.CreateTempFile()
						if err != nil {
							return nil,nil,err
						}

						err = ioutil.WriteFile(imageFile.Name(), frame, os.ModePerm)
						if err != nil {
							return nil,nil,err
						}
						break
					}
				}
			}
			v, err := elem.GetString()
			if err == nil {
				tn, err := dicomtag.FindTagInfo(elem.Tag)
				if err == nil {
					mapping[tn.Name] = v
				}
			}
		}
	}

	if imageFile != nil {
		defer common.FileDelete(imageFile.Name())
	}

	return &mapping,nil,err
}

func (defaultIndexer *DefaultIndexer) Index(path string,options *Options) (string, *Mapping, *[]byte,error) {
	var err error
	var mimeType string
	var mapping *Mapping
	var thumbnail *[]byte
	var buffer []byte

	s,err := common.FileSize(path)
	if err != nil {
		return mimeType, mapping, thumbnail,err
	}

	readComplete := s < 1024 * 1024

	if readComplete {
		buffer,err = ioutil.ReadFile(path)
	} else {
		buffer,err = common.ReadHeader(path)
	}
	if err != nil {
		return mimeType, mapping, thumbnail,err
	}

	mimeType, _ = common.DetectMimeType(buffer)

	if !readComplete {
		buffer = buffer[0:0]
	}

	if common.IsImageMimeType(mimeType) {
		txt,orientation,err := tools.Ocr(path)
		if err != nil {
			return mimeType, mapping, thumbnail,err
		}

		fmt.Println(txt)
		fmt.Println(orientation)

	} else {
		switch mimeType {
		case common.MIMETYPE_APPLICATION_DICOM.MimeType:
			mapping, thumbnail, err = defaultIndexer.indexDicom(path, buffer,options)
		}
	}

	return mimeType, mapping, thumbnail,err
}
