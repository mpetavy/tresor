package database

import (
	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/cache"
	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/errors"
	"net/http"
	"net/url"
)

//go:generate

const (
	TYPE        = "db"
	TYPE_STORM  = "storm"
	TYPE_BADGER = "badger"
	TYPE_MONGO  = "mongo"
	TYPE_PGSQL  = "pgsql"

	QUERY = "query"
)

type Options struct {
}

type Database interface {
	Init(*common.Jason) error
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

type instance struct {
	cfg  *common.Jason
	pool chan Database
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason, router *mux.Router) error {
	pool := make(chan Database, 10)
	for i := 0; i < 10; i++ {
		db, err := create(cfg)
		if common.Fatal(err) {
			return err
		}

		pool <- db
	}

	common.Info("Registered database '%s'", name)

	instances[name] = instance{cfg, pool}

	router.PathPrefix("/"+name).Subrouter().HandleFunc("/{sql}", func(rw http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)

		sql, err := url.QueryUnescape(v["sql"])
		if common.Error(err) {
			return
		}

		var ba []byte

		c, ok := cache.Get(QUERY, sql)

		if ok {
			ba = c.([]byte)
		}

		force := common.ToBool(r.URL.Query().Get("force"))

		if force || ba == nil {
			common.Debug(sql)

			common.Error(Exec(name, func(db Database) error {
				result, err := db.SQL(sql)
				if common.Error(err) {
					return err
				}

				ba = []byte(result)

				cache.Put(QUERY, sql, ba)

				return nil
			}))
		}

		//rw.Header().Add("Content-type", "application/json")
		rw.Write(ba)

		return
	})

	rebuild, err := cfg.Bool("rebuild")
	if common.Error(err) {
		return err
	}

	if rebuild {
		common.Info("Create Schema %s", name)

		common.Error(Exec(name, func(db Database) error {
			err := db.CreateSchema([]interface{}{&models.User{}, &models.Bucket{}})
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

func Get(name string) Database {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	db := <-i.pool

	return db
}

func Put(name string, db Database) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	i.pool <- db
}

func Exec(name string, fn func(db Database) error) error {
	db := Get(name)
	defer Put(name, db)

	return fn(db)
}

func create(cfg *common.Jason) (Database, error) {
	driver, err := cfg.String("driver")
	if common.Error(err) {
		return nil, err
	}

	var db Database

	switch driver {
	case TYPE_STORM:
		db, err = NewStormDB()
	case TYPE_BADGER:
		db, err = NewBadgerDB()
	case TYPE_MONGO:
		db, err = NewMongoDB()
	case TYPE_PGSQL:
		db, err = NewPgsqlDB()
	default:
		return nil, &errors.ErrUnknownDriver{driver}
	}

	if common.Error(err) {
		return nil, err
	}

	err = db.Init(cfg)
	if common.Error(err) {
		return nil, err
	}

	err = db.Start()
	if common.Error(err) {
		return nil, err
	}

	return db, nil
}
