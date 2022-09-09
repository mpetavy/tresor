package index

import (
	"github.com/mpetavy/tresor/utils"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/errors"
)

type Cfg struct {
	Driver string `json:"driver" html:"Driver"`
}

type Options struct {
}

type Mapping map[string]string

type Handle interface {
	Init(*Cfg) error
	Start() error
	Stop() error
	Index(path string, options *Options) (string, Mapping, []byte, string, utils.Orientation, error)
}

type IndexResult struct {
	MimeType    string
	Mapping     Mapping
	Thumbnail   []byte
	Fulltext    string
	Orientation utils.Orientation
}

var (
	cfg  *Cfg
	pool chan Handle
)

func Init(c *Cfg, router *mux.Router) error {
	cfg = c

	pool = make(chan Handle, 10)
	for i := 0; i < 10; i++ {
		handle, err := create(cfg)
		if common.Error(err) {
			common.Error(err)
		}

		pool <- handle
	}

	common.Info("Service index started")

	return nil
}

func Close() {
	close(pool)
	for handle := range pool {
		common.Error(handle.Stop())
	}

	common.Info("Service index stopped")
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

func create(cfg *Cfg) (Handle, error) {
	var handle Handle
	var err error

	switch cfg.Driver {
	case DEFAULT_INDEXER:
		handle, err = NewDefaultIndexer()
		if common.Error(err) {
			return nil, err
		}
	default:
		return nil, &errors.ErrUnknownDriver{Driver: cfg.Driver}
	}

	err = handle.Init(cfg)
	if common.Error(err) {
		return nil, err
	}

	err = handle.Start()
	if common.Error(err) {
		return nil, err
	}

	return handle, nil
}
