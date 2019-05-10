package storage

import (
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
	Load(string, io.Writer, *Options) (*[]byte, int64, error)
	Delete(string, *Options) error
}

type instance struct {
	cfg  *common.Jason
	pool *chan *Storage
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason, router *mux.Router) error {
	pool := make(chan *Storage, 10)
	for i := 0; i < 10; i++ {
		db, err := create(cfg)
		if err != nil {
			common.Fatal(err)
		}

		pool <- db
	}

	instances[name] = instance{cfg, &pool}

	router.PathPrefix("/"+name).Subrouter().HandleFunc("/{uid}", func(rw http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)

		storage := Get(name)
		defer Put(name, storage)

		_, _, err := (*storage).Load(v["uid"], rw, nil)
		if err != nil {
			common.Error(err)

			return
		}
	})

	common.Info("Registered storage '%s'", name)

	rebuild, err := cfg.Bool("rebuild")
	if err != nil {
		return err
	}

	if rebuild {
		start := time.Now()

		common.Info("Rebuild started ...")

		storage := Get(name)

		var c int

		c, err = (*storage).Rebuild()
		if err != nil {
			return err
		}

		common.Info("Rebuild successfully completed. time needed %v, %d buckets", time.Now().Sub(start), c)
	}

	return nil
}

func Close() {
}

func Get(name string) *Storage {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	storage := <-*(i.pool)

	return storage
}

func Put(name string, storage *Storage) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	*(i.pool) <- storage
}

func Exec(name string, fn func(storage *Storage) error) error {
	storage := Get(name)
	defer Put(name, storage)

	return fn(storage)
}

func create(cfg *common.Jason) (*Storage, error) {
	driver, err := cfg.String("driver")
	if err != nil {
		return nil, err
	}

	var storage Storage

	switch driver {
	case TYPE_FS:
		storage, err = NewFs()
		if err != nil {
			return nil, err
		}
	case TYPE_SHA:
		storage, err = NewSha()
		if err != nil {
			return nil, err
		}
	default:
		return nil, &errors.ErrUnknownDriver{driver}
	}

	err = storage.Init(cfg)
	if err != nil {
		return nil, err
	}

	err = storage.Start()
	if err != nil {
		return nil, err
	}

	return &storage, nil
}
