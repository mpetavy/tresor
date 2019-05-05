package index

import (
	"github.com/mpetavy/common"
	"github.com/mpetavy/go-dicom"
	"github.com/mpetavy/go-dicom/dicomtag"
)

const (
	DEFAULT_MAPPER = ""
)

type DefaultMapper struct {
	name string
}

func NewDefaultMapper() (*DefaultMapper, error) {
	return &DefaultMapper{}, nil
}

func (defaultMapper *DefaultMapper) Init(cfg *common.Jason) error {
	name, err := cfg.String("name")
	if err != nil {
		return err
	}
	defaultMapper.name = name

	return nil
}

func (defaultMapper *DefaultMapper) Start() error {
	return nil
}

func (defaultMapper *DefaultMapper) Stop() error {
	return nil
}

func (defaultMapper *DefaultMapper) Index(buffer *[]byte, options *Options) (string, *Mapping, error) {
	var m Mapping

	mimeType, _ := common.DetectMimeType(*buffer)

	if mimeType == "application/dicom" {
		m = make(Mapping)
		dataset, err := dicom.ReadDataSetInBytes(*buffer, dicom.ReadOptions{DropPixelData: true})
		if err == nil {
			for _, elem := range dataset.Elements {
				v, err := elem.GetString()
				if err == nil {
					tn, err := dicomtag.FindTagInfo(elem.Tag)
					if err == nil {
						m[tn.Name] = v
					}
				}
			}
		}
	}

	return mimeType, &m, nil
}
