package index

import (
	"net/http"

	"github.com/mpetavy/tresor/utils"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/errors"
)

const (
	TYPE = "index"
)

type Options struct {
}

type Mapping map[string]string

type Index interface {
	Init(*common.Jason) error
	Start() error
	Stop() error
	Index(path string, options *Options) (string, Mapping, *[]byte, string, utils.Orientation, error)
}

type instance struct {
	cfg  *common.Jason
	pool chan Index
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason, router *mux.Router) error {
	pool := make(chan Index, 10)
	for i := 0; i < 10; i++ {
		index, err := create(cfg)
		if common.Error(err) {
			common.Fatal(err)
		}

		pool <- index
	}

	instances[name] = instance{cfg, pool}

	router.PathPrefix("/"+name).Subrouter().HandleFunc("/{uid}", func(rw http.ResponseWriter, r *http.Request) {
		//v := mux.Vars(r)
		//
		//index := Get(name)
		//defer Put(name, index)
		//
		//_, _, err := (*index).Load(v["uid"], rw, nil)
		//if common.Error(err) {
		//	common.Error(err)
		//
		//	return
		//}
	})

	common.Info("Registered index '%s'", name)

	return nil
}

func Close() {
}

func Get(name string) Index {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	index := <-i.pool

	return index
}

func Put(name string, index Index) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	i.pool <- index
}

func Exec(name string, fn func(index Index) error) error {
	index := Get(name)
	defer Put(name, index)

	return fn(index)
}

func create(cfg *common.Jason) (Index, error) {
	driver, err := cfg.String("driver")
	if common.Error(err) {
		return nil, err
	}

	var index Index

	switch driver {
	case DEFAULT_INDEXER:
		index, err = NewDefaultIndexer()
		if common.Error(err) {
			return nil, err
		}
	default:
		return nil, &errors.ErrUnknownDriver{driver}
	}

	err = index.Init(cfg)
	if common.Error(err) {
		return nil, err
	}

	err = index.Start()
	if common.Error(err) {
		return nil, err
	}

	return index, nil
}
