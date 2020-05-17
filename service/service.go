package service

import (
	"fmt"
	"github.com/mpetavy/tresor/service/errors"
	"github.com/mpetavy/tresor/service/index"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/service/database"
	"github.com/mpetavy/tresor/service/storage"
)

func InitServices(router *mux.Router) error {
	cwd, err := os.Getwd()
	if common.Error(err) {
		return err
	}

	f, err := os.Open(filepath.Join(cwd, "tresor.json"))
	if common.Error(err) {
		return err
	}
	defer func() {
		common.Error(f.Close())
	}()

	b, err := ioutil.ReadAll(f)
	if common.Error(err) {
		return err
	}

	serverCfg, err := common.NewJason(string(b))
	if common.Error(err) {
		return err
	}

	if !serverCfg.Exists("services") {
		return fmt.Errorf("Undefined services")
	}

	for i := 0; i < serverCfg.ArrayCount("services"); i++ {
		serviceCfg, err := serverCfg.Array("services", i)
		if common.Error(err) {
			return err
		}

		serviceName, err := serviceCfg.String("name")
		if common.Error(err) {
			return err
		}
		serviceType, err := serviceCfg.String("type")
		if common.Error(err) {
			return err
		}

		switch serviceType {
		case storage.TYPE:
			err := storage.Init(serviceName, serviceCfg, router)
			if common.Error(err) {
				return err
			}
		case database.TYPE:
			err := database.Init(serviceName, serviceCfg, router)
			if common.Error(err) {
				return err
			}
		case index.TYPE:
			err := index.Init(serviceName, serviceCfg, router)
			if common.Error(err) {
				return err
			}
		default:
			return &errors.ErrUnknownService{serviceType}
		}
	}

	return nil
}

func StopServices() error {
	common.DebugFunc()

	storage.Close()
	database.Close()
	index.Close()

	return nil
}
