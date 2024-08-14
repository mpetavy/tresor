package index

import (
	"bytes"
	"fmt"
	"github.com/dsoprea/go-exif/v3"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"os"

	"github.com/mpetavy/tresor/utils"

	"github.com/mpetavy/common"
	"github.com/mpetavy/go-dicom"
	"github.com/mpetavy/go-dicom/dicomtag"
)

const (
	DEFAULT_INDEXER = "default"
	EXIF_PREFIX     = "EXIF"
	DICOM_PREFIX    = "DCM"
)

type DefaultIndexer struct {
}

func NewDefaultIndexer() (*DefaultIndexer, error) {
	return &DefaultIndexer{}, nil
}

func (defaultIndexer *DefaultIndexer) Init(cfg *Cfg) error {
	return nil
}

func (defaultIndexer *DefaultIndexer) Start() error {
	return nil
}

func (defaultIndexer *DefaultIndexer) Stop() error {
	return nil
}

func (defaultIndexer *DefaultIndexer) indexPDF(path string, buffer []byte, options *Options) (Mapping, []byte, string, error) {
	mapping := make(Mapping)

	pdfReader, err := model.NewPdfReader(bytes.NewReader(buffer))
	if common.Error(err) {
		return mapping, nil, "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if common.Error(err) {
		return mapping, nil, "", err
	}

	var strbuf bytes.Buffer

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if common.Error(err) {
			return mapping, nil, "", err
		}

		ex, err := extractor.New(page)
		if common.Error(err) {
			return mapping, nil, "", err
		}

		text, err := ex.ExtractText()
		if common.Error(err) {
			return mapping, nil, "", err
		}

		strbuf.WriteString(text)
	}

	return mapping, nil, strbuf.String(), err
}

func (defaultIndexer *DefaultIndexer) indexExif(path string, buffer []byte, options *Options) (Mapping, []byte, error) {
	exifData, err := exif.SearchAndExtractExif(buffer)
	if common.WarnError(err) {
		return nil, nil, err
	}

	exifTags, _, err := exif.GetFlatExifData(exifData, &exif.ScanOptions{})
	if common.Error(err) {
		return nil, nil, err
	}

	mapping := make(Mapping)
	for _, tag := range exifTags {
		if tag.Formatted == "" {
			continue
		}

		mapping[fmt.Sprintf("%s.%s", EXIF_PREFIX, tag.TagName)] = tag.Formatted
	}

	return mapping, nil, err
}

func (defaultIndexer *DefaultIndexer) indexDicom(path string, buffer []byte, options *Options) (Mapping, []byte, error) {
	mapping := make(Mapping)

	var imageFile *os.File
	var dataset *dicom.DataSet

	if len(buffer) > 0 {
		var err error

		dataset, err = dicom.ReadDataSetInBytes(buffer, dicom.ReadOptions{})
		if common.Error(err) {
			return nil, nil, err
		}
	} else {
		var err error

		dataset, err = dicom.ReadDataSetFromFile(path, dicom.ReadOptions{})
		if common.Error(err) {
			return nil, nil, err
		}
	}

	var representativeFrameNumber uint16

	elem, err := dataset.FindElementByTag(dicomtag.RepresentativeFrameNumber)
	if err == nil {
		representativeFrameNumber, err = elem.GetUInt16()
		if common.WarnError(err) {
			return nil, nil, err
		}
	}

	for _, elem := range dataset.Elements {
		if elem.Tag == dicomtag.PixelData {
			data := elem.Value[0].(dicom.PixelDataInfo)
			for i, frame := range data.Frames {
				if uint16(i) == representativeFrameNumber {
					var err error

					imageFile, err = common.CreateTempFile()
					if common.Error(err) {
						return nil, nil, err
					}

					err = os.WriteFile(imageFile.Name(), frame, common.DefaultFileMode)
					if common.Error(err) {
						return nil, nil, err
					}

					break
				}
			}
		}
		v, err := elem.GetString()
		if err == nil {
			tn, err := dicomtag.FindTagInfo(elem.Tag)
			if err == nil && v != "" {
				mapping[fmt.Sprint("%s.%s", DICOM_PREFIX, tn.Name)] = v
			}
		}
	}

	if imageFile != nil {
		defer func() {
			common.Error(common.FileDelete(imageFile.Name()))
		}()
	}

	return mapping, nil, nil
}

func (defaultIndexer *DefaultIndexer) Index(path string, options *Options) (string, Mapping, []byte, string, utils.Orientation, error) {
	var err error
	var mimeType string
	var mapping Mapping
	var thumbnail []byte
	var buffer []byte
	var fulltext string
	var orientation utils.Orientation

	s, err := common.FileSize(path)
	if common.Error(err) {
		return mimeType, mapping, thumbnail, fulltext, orientation, err
	}

	readComplete := s < 1024*1024

	if readComplete {
		buffer, err = os.ReadFile(path)
	} else {
		buffer, err = common.ReadFileHeader(path)
	}
	if common.Error(err) {
		return mimeType, mapping, thumbnail, fulltext, orientation, err
	}

	mt, err := common.DetectMimeType(path, buffer)
	if common.Error(err) {
		return mimeType, mapping, thumbnail, fulltext, orientation, err
	}

	mimeType = mt.MimeType

	if !readComplete {
		buffer = buffer[0:0]
	}

	switch mimeType {
	case common.MimetypeApplicationPdf.MimeType:
		mapping, thumbnail, fulltext, err = defaultIndexer.indexPDF(path, buffer, options)
	case common.MimetypeApplicationDicom.MimeType:
		mapping, thumbnail, err = defaultIndexer.indexDicom(path, buffer, options)
	default:
		if common.IsImageMimeType(mimeType) {
			if mimeType == common.MimetypeImageJpeg.MimeType || mimeType == common.MimetypeImageTiff.MimeType {
				mapping, thumbnail, err = defaultIndexer.indexExif(path, buffer, options)
				common.DebugError(err)
			}

			fulltext, orientation, err = utils.Ocr(path)
			common.DebugError(err)
		}
	}

	return mimeType, mapping, thumbnail, fulltext, orientation, err
}
