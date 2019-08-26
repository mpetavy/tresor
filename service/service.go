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
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(cwd, "tresor.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	serverCfg, err := common.NewJason(string(b))
	if err != nil {
		return err
	}

	if !serverCfg.Exists("services") {
		return fmt.Errorf("Undefined services")
	}

	for i := 0; i < serverCfg.ArrayCount("services"); i++ {
		serviceCfg, err := serverCfg.Array("services", i)
		if err != nil {
			return err
		}

		serviceName, err := serviceCfg.String("name")
		if err != nil {
			return err
		}
		serviceType, err := serviceCfg.String("type")
		if err != nil {
			return err
		}

		switch serviceType {
		case storage.TYPE:
			err := storage.Init(serviceName, serviceCfg, router)
			if err != nil {
				return err
			}
		case database.TYPE:
			err := database.Init(serviceName, serviceCfg)
			if err != nil {
				return err
			}
		case index.TYPE:
			err := index.Init(serviceName, serviceCfg, router)
			if err != nil {
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

	return nil
}
