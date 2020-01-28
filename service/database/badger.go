package database

import (
	"github.com/dgraph-io/badger"
	"github.com/mpetavy/common"
)

type BadgerDB struct {
	Path string
	DB   *badger.DB
}

func NewBadgerDB() (*BadgerDB, error) {
	return &BadgerDB{}, nil
}

func (db *BadgerDB) Init(cfg *common.Jason) error {
	path, err := cfg.String("path")
	if common.Error(err) {
		return err
	}

	db.Path = common.CleanPath(path)

	return nil
}

func (db *BadgerDB) CreateSchema([]interface{}) error {
	return nil
}

func (db *BadgerDB) SwitchIndices(models []interface{}, enable bool) error {
	return nil
}

func (db *BadgerDB) Query(rows interface{}, sql string) (string, error) {
	return "", nil
}

func (db *BadgerDB) Start() error {
	opts := badger.DefaultOptions
	opts.Dir = db.Path
	opts.ValueDir = db.Path
	x, err := badger.Open(opts)
	db.DB = x

	return err
}

func (db *BadgerDB) Stop() error {
	var err error

	if db.DB != nil {
		err = db.DB.Close()
	}

	return err
}
