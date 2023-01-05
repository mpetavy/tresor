package service

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/database"
	"github.com/mpetavy/tresor/service/index"
	"github.com/mpetavy/tresor/service/storage"
)

type TresorCfg struct {
	common.Configuration
	Database database.Cfg `json:"database" html:"Database"`
	Index    index.Cfg    `json:"index" html:"Index"`
	Storage  storage.Cfg  `json:"storage" html:"Storage"`
}

func StartServices(router *mux.Router) error {
	cfg := &TresorCfg{}

	ba, err := common.LoadConfigurationFile()
	if common.Error(err) {
		return err
	}

	err = json.Unmarshal(ba, cfg)
	if common.Error(err) {
		return err
	}

	err = database.Init(&cfg.Database, router)
	if common.Error(err) {
		return err
	}

	err = index.Init(&cfg.Index, router)
	if common.Error(err) {
		return err
	}

	err = storage.Init(&cfg.Storage, router)
	if common.Error(err) {
		return err
	}

	return nil
}

func StopServices() error {
	common.DebugFunc()

	storage.Close()
	index.Close()
	database.Close()

	return nil
}
