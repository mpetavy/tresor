package database

import (
	"context"
	"github.com/mpetavy/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDB struct {
	Name           string
	URL            string
	Client         *mongo.Client
	Ctx            context.Context
	ConnectTimeout int
}

func NewMongoDB() (*MongoDB, error) {
	return &MongoDB{}, nil
}

func (db *MongoDB) Init(cfg *common.Jason) error {
	name, err := cfg.String("name")
	if err != nil {
		return err
	}

	db.Name = name

	url, err := cfg.String("url")
	if err != nil {
		return err
	}

	db.URL = url

	ti, err := cfg.Int("connectTimeout", 3000)
	if err != nil {
		return err
	}

	db.ConnectTimeout = ti

	return nil
}

func (db *MongoDB) CreateSchema([]interface{}) error {
	return nil
}

func (db *MongoDB) SwitchIndices(models []interface{}, enable bool) error {
	return nil
}

func (db *MongoDB) Start() error {
	var err error

	db.Ctx, _ = context.WithTimeout(context.Background(), time.Duration(db.ConnectTimeout)*time.Millisecond)
	db.Client, err = mongo.Connect(db.Ctx, options.Client().ApplyURI(db.URL))
	if err != nil {
		return err
	}

	err = db.Client.Ping(db.Ctx, nil)
	if err != nil {
		return err
	}

	return err
}

func (db *MongoDB) Stop() error {
	return nil
}
