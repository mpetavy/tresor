package database

import (
	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/cache"
	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/errors"
	"net/http"
)

//go:generate

const (
	TYPE_MONGODB = "mongodb"
	TYPE_PGSQL   = "pgsql"

	QUERY = "query"
)

type Cfg struct {
	Driver   string `json:"driver" html:"Driver"`
	Hostname string `json:"hostname" html:"Host name"`
	Port     int    `json:"port" html:"Port" html_min:"0" html_max:"65535"`
	Username string `json:"username" html:"Username"`
	Password string `json:"password" html:"Password"`
	Instance string `json:"instance" html:"Instance"`
	SSL      bool   `json:"ssl" html:"SSL"`
	Rebuild  bool   `json:"rebuild" html:"Rebuild"`
}

type Options struct {
}

type Handle interface {
	Init(*Cfg) error
	Start() error
	Stop() error

	CreateSchema([]interface{}) error
	SwitchIndices(models []interface{}, enable bool) error
	SQL(sql string) (string, error)

	SaveBucket(doc *models.Bucket, options *Options) error
	LoadBucket(field string, value interface{}, doc *models.Bucket, options *Options) error
	DeleteBucket(field string, value interface{}, id int, options *Options) error

	SaveUser(user *models.User, options *Options) error
	LoadUser(field string, value interface{}, user *models.User, options *Options) error
	DeleteUser(field string, value interface{}, id int, options *Options) error
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
			return err
		}

		pool <- handle
	}

	common.Info("Registered database")

	router.PathPrefix("/db/").Handler(http.StripPrefix("/db/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sql := r.URL.Path

		var ba []byte

		c, ok := cache.Get(QUERY, sql)

		if ok {
			ba = c.([]byte)
		}

		force := common.ToBool(r.URL.Query().Get("force"))

		if force || ba == nil {
			common.Debug(sql)

			common.Error(Exec(func(handle Handle) error {
				result, err := handle.SQL(sql)
				if common.Error(err) {
					return err
				}

				ba = []byte(result)

				cache.Put(QUERY, sql, ba)

				return nil
			}))
		}

		//rw.Header().Add("Content-type", "application/json")
		_, err := rw.Write(ba)
		common.Error(err)
	})))

	if cfg.Rebuild {
		common.Info("Create Schema")

		common.Error(Exec(func(handle Handle) error {
			err := handle.CreateSchema([]interface{}{&models.User{}, &models.Bucket{}})
			if common.Error(err) {
				return err
			}

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

func create(cfg *Cfg) (Handle, error) {
	var handle Handle
	var err error

	switch cfg.Driver {
	case TYPE_MONGODB:
		handle, err = NewMongoDB()
	case TYPE_PGSQL:
		handle, err = NewPgsqlDB()
	default:
		return nil, &errors.ErrUnknownDriver{Driver: cfg.Driver}
	}

	if common.Error(err) {
		return nil, err
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
