package database

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/mpetavy/common"
)

type PgsqlDB struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	DB       *pg.DB
}

func NewPgsqlDB() (*PgsqlDB, error) {
	return &PgsqlDB{}, nil
}

func (db *PgsqlDB) Init(cfg *common.Jason) error {
	var err error

	db.Username, err = cfg.String("username")
	if err != nil {
		return err
	}
	db.Password, err = cfg.String("password")
	if err != nil {
		return err
	}
	db.Host, err = cfg.String("host")
	if err != nil {
		return err
	}
	db.Port, err = cfg.Int("port")
	if err != nil {
		return err
	}
	db.Database, err = cfg.String("database")
	if err != nil {
		return err
	}

	return nil
}

func (db *PgsqlDB) CreateSchema(models []interface{}) error {
	for _, model := range models {
		err := db.DB.DropTable(model, &orm.DropTableOptions{})
		if err != nil {
			common.Warn(err.Error())
		}

		err = db.DB.CreateTable(model, &orm.CreateTableOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *PgsqlDB) Start() error {
	db.DB = pg.Connect(&pg.Options{
		User:     db.Username,
		Password: db.Password,
		Addr:     fmt.Sprintf("%s:%d", db.Host, db.Port),
		Database: db.Database,
	})

	return nil
}

func (db *PgsqlDB) Stop() error {
	if db.DB != nil {
		return db.DB.Close()
	}

	return nil
}
