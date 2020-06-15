package storage

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/mpetavy/go-dicom"
	"github.com/mpetavy/go-dicom/dicomtag"
	"github.com/mpetavy/tresor/utils"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/errors"
)

const (
	TYPE  = "storage"
	PAGE  = "page"
	UNZIP = "unzip"
)

type Options struct {
	VolumeName string
}

type VolumeCfg struct {
	Name string `json:"name" html:"Name"`
	Path string `json:"path" html:"path"`
	Flat bool   `json:"flat" html:"Flat"`
	Zip  bool   `json:"zip" html:"Zip"`
}

type Cfg struct {
	Driver  string      `json:"driver" html:"Driver"`
	Rebuild bool        `json:"rebuild" html:"Rebuild"`
	Volumes []VolumeCfg `json:"volumes" html:"Volumes"`
}

type Handle interface {
	Init(*Cfg) error
	Start() error
	Stop() error
	Rebuild() (int, error)
	Store(string, io.Reader, *Options) (string, *[]byte, error)
	Load(string, io.Writer, *Options) (string, *[]byte, int64, error)
	Delete(string, *Options) error
}

var (
	cfg  *Cfg
	pool chan Handle
)

func Init(c *Cfg, router *mux.Router) error {
	cfg = c

	pool = make(chan Handle, 10)
	for i := 0; i < 10; i++ {
		storage, err := create(cfg)
		if common.Error(err) {
			common.Error(err)
		}

		pool <- storage
	}

	router.PathPrefix("/" + TYPE + "/").Handler(http.StripPrefix("/"+TYPE+"/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		uid := r.URL.Path

		common.Error(Exec(func(storage Handle) error {
			_, _, _, err := storage.Load(uid, rw, nil)
			if common.Error(err) {
				return err
			}

			return nil
		}))
	})))

	router.PathPrefix("/" + TYPE + "-pixeldata/").Handler(http.StripPrefix("/"+TYPE+"-pixeldata/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		uid := r.URL.Path

		common.Error(Exec(func(storage Handle) error {
			var ba []byte
			var buf bytes.Buffer

			_, _, _, err := storage.Load(uid, &buf, nil)
			if common.Error(err) {
				return err
			}

			ba = buf.Bytes()

			mimeType := common.DetectMimeType("", ba).MimeType

			if mimeType == common.MimetypeApplicationDicom.MimeType {
				dcm, err := dicom.ReadDataSetInBytes(buf.Bytes(), dicom.ReadOptions{DropPixelData: false})
				if common.Error(err) {
					return err
				}

				for _, elem := range dcm.Elements {
					if elem.Tag == dicomtag.PixelData {
						data := elem.Value[0].(dicom.PixelDataInfo)
						ba = data.Frames[0]

						mimeType = common.DetectMimeType("", ba).MimeType

						break
					}
				}
			}

			if ba == nil {
				return fmt.Errorf("cannot handle content with mimeType %s", mimeType)
			}

			if common.IsImageMimeType(mimeType) {
				if mimeType != common.MimetypeImageJpeg.MimeType {
					img, err := utils.LoadImage(ba)
					if common.Error(err) {
						return err
					}

					if mimeType != common.MimetypeImageJpeg.MimeType {
						err = jpeg.Encode(&buf, img, &jpeg.Options{80})
						if common.Error(err) {
							return err
						}

						ba = buf.Bytes()
					}
				}

				rw.Header().Add("Content-type", common.MimetypeImageJpeg.MimeType)
				rw.Header().Add("Content-length", strconv.Itoa(len(ba)))

				_, err = io.Copy(rw, bytes.NewReader(ba))
				if common.Error(err) {
					return err
				}
			}

			return nil
		}))
	})))

	common.Info("Registered storage")

	if cfg.Rebuild {
		start := time.Now()

		common.Info("Rebuild started ...")

		common.Error(Exec(func(storage Handle) error {
			c, err := storage.Rebuild()
			if common.Error(err) {
				return err
			}

			common.Info("Rebuild successfully completed. time needed %v, %d buckets", time.Since(start), c)

			return nil
		}))
	}

	return nil
}

func Close() {
}

func Get() Handle {
	handle := <-pool

	return handle
}

func Put(handle Handle) {
	pool <- handle
}

func Exec(fn func(handle Handle) error) error {
	handle := Get()
	defer Put(handle)

	return fn(handle)
}

func getFromList(l list.List, index int) interface{} {
	e := l.Front()

	for i := 0; i < index; i++ {
		e = e.Next()
	}

	return e.Value
}

func create(cfg *Cfg) (Handle, error) {
	var storage Handle
	var err error

	switch cfg.Driver {
	case TYPE_FS:
		storage, err = NewFs()
		if common.Error(err) {
			return nil, err
		}
	case TYPE_SHA:
		storage, err = NewSha()
		if common.Error(err) {
			return nil, err
		}
	default:
		return nil, &errors.ErrUnknownDriver{Driver: cfg.Driver}
	}

	err = storage.Init(cfg)
	if common.Error(err) {
		return nil, err
	}

	err = storage.Start()
	if common.Error(err) {
		return nil, err
	}

	return storage, nil
}
