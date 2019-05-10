package database

import (
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/errors"
)

//go:generate

const (
	TYPE        = "db"
	TYPE_STORM  = "storm"
	TYPE_BADGER = "badger"
	TYPE_MONGO  = "mongo"
	TYPE_PGSQL  = "pgsql"
)

type Options struct {
}

type Database interface {
	Init(*common.Jason) error
	Start() error
	Stop() error

	CreateSchema([]interface{}) error
	SwitchIndices(models []interface{}, enable bool) error

	SaveBucket(doc *models.Bucket, options *Options) error
	LoadBucket(field string, value interface{}, doc *models.Bucket, options *Options) error
	DeleteBucket(field string, value interface{}, id int, options *Options) error

	SaveUser(user *models.User, options *Options) error
	LoadUser(field string, value interface{}, user *models.User, options *Options) error
	DeleteUser(field string, value interface{}, id int, options *Options) error
}

type instance struct {
	cfg  *common.Jason
	pool *chan *Database
}

var instances map[string]instance

func init() {
	instances = make(map[string]instance)
}

func Init(name string, cfg *common.Jason) error {
	pool := make(chan *Database, 10)
	for i := 0; i < 10; i++ {
		db, err := create(cfg)
		if err != nil {
			common.Fatal(err)
		}

		pool <- db
	}

	common.Info("Registered database '%s'", name)

	instances[name] = instance{cfg, &pool}

	common.Info("Create Schema %s", name)

	db := Get(name)
	defer Put(name, db)

	return (*db).CreateSchema([]interface{}{&models.User{}, &models.Bucket{}})
}

func Close() {
}

func Get(name string) *Database {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	db := <-*(i.pool)

	return db
}

func Put(name string, db *Database) {
	i, ok := instances[name]

	if !ok {
		common.Fatal(&errors.ErrUnknownService{name})
	}

	*(i.pool) <- db
}

func Exec(name string, fn func(db *Database) error) error {
	db := Get(name)
	defer Put(name, db)

	return fn(db)
}

func create(cfg *common.Jason) (*Database, error) {
	driver, err := cfg.String("driver")
	if err != nil {
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

	if err != nil {
		return nil, err
	}

	err = db.Init(cfg)
	if err != nil {
		return nil, err
	}

	err = db.Start()
	if err != nil {
		return nil, err
	}

	return &db, nil
}
