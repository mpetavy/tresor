package database

import (
	"context"
	"fmt"
	"github.com/mpetavy/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDB struct {
	Name           string
	URL            string
	ConnectTimeout int
	Client         *mongo.Client
	Ctx            context.Context
}

func NewMongoDB() (*MongoDB, error) {
	return &MongoDB{}, nil
}

func (db *MongoDB) Init(cfg *Cfg) error {
	db.Name = cfg.Instance
	db.URL = fmt.Sprintf("mongodb://%s:%d/?readPreference=primary&appname=%s&ssl=%v", cfg.Hostname, cfg.Port, common.Title(), cfg.SSL)
	db.ConnectTimeout = 3000

	return nil
}

func (db *MongoDB) CreateSchema([]interface{}) error {
	db.Client.Database(db.Name).Drop(nil)

	return nil
}

func (db *MongoDB) SwitchIndices(models []interface{}, enable bool) error {
	return nil
}

func (db *MongoDB) SQL(sql string) (string, error) {
	return "", nil
}

func (db *MongoDB) Start() error {
	var err error

	db.Ctx, _ = context.WithTimeout(context.Background(), time.Duration(db.ConnectTimeout)*time.Millisecond)
	db.Client, err = mongo.Connect(db.Ctx, options.Client().ApplyURI(db.URL))
	if common.Error(err) {
		return err
	}

	err = db.Client.Ping(db.Ctx, nil)
	if common.Error(err) {
		return err
	}

	return err
}

func (db *MongoDB) Stop() error {
	return nil
}
