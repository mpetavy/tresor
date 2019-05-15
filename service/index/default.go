package index

import (
	"github.com/mpetavy/common"
	"github.com/mpetavy/go-dicom"
	"github.com/mpetavy/go-dicom/dicomtag"
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

func (defaultIndexer *DefaultIndexer) indexDicom(path string,options *Options) (*Mapping, *[]byte,error) {
	mapping := make(Mapping)

	var imageFile *os.File

	dataset, err := dicom.ReadDataSetFromFile(path, dicom.ReadOptions{})
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

	header,err := common.ReadHeader(path)
	if err != nil {
		return mimeType, mapping, thumbnail,err
	}

	mimeType, _ = common.DetectMimeType(header)

	if common.IsImageMimeType(mimeType) {
	} else {
		if mimeType == "application/dicom" {
			mapping, thumbnail, err = defaultIndexer.indexDicom(path, options)
		}
	}

	return mimeType, mapping, thumbnail,err
}
