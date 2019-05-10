package database

import (
	"github.com/asdine/storm"
	"github.com/mpetavy/common"
)

type StormDB struct {
	Path string
	DB   *storm.DB
}

func NewStormDB() (*StormDB, error) {
	return &StormDB{}, nil
}

func (db *StormDB) Init(cfg *common.Jason) error {
	path, err := cfg.String("path")
	if err != nil {
		return err
	}

	db.Path = common.CleanPath(path)

	return nil
}

func (db *StormDB) CreateSchema([]interface{}) error {
	return nil
}

func (db *StormDB) SwitchIndices(models []interface{}, enable bool) error {
	return nil
}

func (db *StormDB) Start() error {
	x, err := storm.Open(db.Path)
	db.DB = x

	return err
}

func (db *StormDB) Stop() error {
	var err error

	if db.DB != nil {
		err = db.DB.Close()
	}

	return err
}
