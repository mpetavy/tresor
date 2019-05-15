package index

import (
	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/errors"
	"net/http"
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
	Index(file string, options *Options) (string, *Mapping, *[]byte, error)
}

type instance struct {
	cfg  *common.Jason
	pool *chan *Index
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason, router *mux.Router) error {
	pool := make(chan *Index, 10)
	for i := 0; i < 10; i++ {
		h, err := create(cfg)
		if err != nil {
			common.Fatal(err)
		}

		pool <- h
	}

	instances[name] = instance{cfg, &pool}

	router.PathPrefix("/"+name).Subrouter().HandleFunc("/{uid}", func(rw http.ResponseWriter, r *http.Request) {
		//v := mux.Vars(r)
		//
		//index := Get(name)
		//defer Put(name, index)
		//
		//_, _, err := (*index).Load(v["uid"], rw, nil)
		//if err != nil {
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

func Get(name string) *Index {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	index := <-*(i.pool)

	return index
}

func Put(name string, index *Index) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	*(i.pool) <- index
}

func Exec(name string, fn func(index *Index) error) error {
	index := Get(name)
	defer Put(name, index)

	return fn(index)
}

func create(cfg *common.Jason) (*Index, error) {
	driver, err := cfg.String("driver")
	if err != nil {
		return nil, err
	}

	var index Index

	switch driver {
	case DEFAULT_INDEXER:
		index, err = NewDefaultIndexer()
		if err != nil {
			return nil, err
		}
	default:
		return nil, &errors.ErrUnknownDriver{driver}
	}

	err = index.Init(cfg)
	if err != nil {
		return nil, err
	}

	err = index.Start()
	if err != nil {
		return nil, err
	}

	return &index, nil
}
