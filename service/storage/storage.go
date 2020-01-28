package storage

import (
	"container/list"
	"io"
	"net/http"
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

type Storage interface {
	Init(*common.Jason) error
	Start() error
	Stop() error
	Rebuild() (int, error)
	Store(string, io.Reader, *Options) (string, *[]byte, error)
	Load(string, io.Writer, *Options) (string, *[]byte, int64, error)
	Delete(string, *Options) error
}

type instance struct {
	cfg  *common.Jason
	pool chan Storage
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason, router *mux.Router) error {
	pool := make(chan Storage, 10)
	for i := 0; i < 10; i++ {
		storage, err := create(cfg)
		if common.Error(err) {
			common.Fatal(err)
		}

		pool <- storage
	}

	instances[name] = instance{cfg, pool}

	router.PathPrefix("/"+name).Subrouter().HandleFunc("/{uid}", func(rw http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)

		common.Error(Exec(name, func(storage Storage) error {
			_, _, _, err := storage.Load(v["uid"], rw, nil)
			if common.Error(err) {
				return err
			}

			return nil
		}))
	})

	common.Info("Registered storage '%s'", name)

	rebuild, err := cfg.Bool("rebuild")
	if common.Error(err) {
		return err
	}

	if rebuild {
		start := time.Now()

		common.Info("Rebuild started ...")

		common.Error(Exec(name, func(storage Storage) error {
			var c int

			c, err = storage.Rebuild()
			if common.Error(err) {
				return err
			}

			common.Info("Rebuild successfully completed. time needed %v, %d buckets", time.Now().Sub(start), c)

			return nil
		}))
	}

	return nil
}

func Close() {
}

func Get(name string) Storage {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	storage := <-i.pool

	return storage
}

func Put(name string, storage Storage) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	i.pool <- storage
}

func Exec(name string, fn func(storage Storage) error) error {
	storage := Get(name)
	defer Put(name, storage)

	return fn(storage)
}

func getFromList(l list.List, index int) interface{} {
	e := l.Front()

	for i := 0; i < index; i++ {
		e = e.Next()
	}

	return e.Value
}

func create(cfg *common.Jason) (Storage, error) {
	driver, err := cfg.String("driver")
	if common.Error(err) {
		return nil, err
	}

	var storage Storage

	switch driver {
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
		return nil, &errors.ErrUnknownDriver{driver}
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
