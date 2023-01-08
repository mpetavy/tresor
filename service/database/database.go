package database

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/cache"
	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:generate

const (
	TYPE_MONGODB = "mongodb"
	TYPE_PGSQL   = "pgsql"

	QUERY = "query"
)

type Cfg struct {
	Driver   string `json:"driver" html:"Driver"`
	Hostname string `json:"hostname" html:"Host name"`
	Port     int    `json:"port" html:"Port" html_min:"0" html_max:"65535"`
	Username string `json:"username" html:"Username"`
	Password string `json:"password" html:"Password"`
	Instance string `json:"instance" html:"Instance"`
	SSL      bool   `json:"ssl" html:"SSL"`
	Rebuild  bool   `json:"rebuild" html:"Rebuild"`
}

type Options struct {
}

type Handle interface {
	Init(*Cfg) error
	Start() error
	Stop() error

	CreateSchema([]interface{}) error
	EnableIndices(models []interface{}, enable bool) error
	SQL(sql string) (string, error)

	SaveBucket(doc *models.Bucket, options *Options) error
	LoadBucket(field string, value interface{}, doc *models.Bucket, options *Options) error
	DeleteBucket(field string, value interface{}, id int, options *Options) error

	SaveUser(user *models.User, options *Options) error
	LoadUser(field string, value interface{}, user *models.User, options *Options) error
	DeleteUser(field string, value interface{}, id int, options *Options) error
}

var (
	cfg  *Cfg
	pool chan Handle
)

func Init(c *Cfg, router *mux.Router) error {
	cfg = c

	pool = make(chan Handle, 10)
	for i := 0; i < 10; i++ {
		handle, err := create(cfg)
		if common.Error(err) {
			return err
		}

		pool <- handle
	}

	common.Info("Service database started")

	router.PathPrefix("/db/").Handler(http.StripPrefix("/db/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		sql := r.URL.Path

		var ba []byte

		c, ok := cache.Get(QUERY, sql)

		if ok {
			ba = c.([]byte)
		}

		force := common.ToBool(r.URL.Query().Get("force"))

		if force || ba == nil {
			common.Debug(sql)

			common.Error(Exec(func(handle Handle) error {
				result, err := handle.SQL(sql)
				if common.Error(err) {
					return err
				}

				ba = []byte(result)

				cache.Put(QUERY, sql, ba)

				return nil
			}))
		}

		_, err := rw.Write(ba)
		common.Error(err)
	})))

	if cfg.Rebuild {
		common.Info("Create Schema")

		err := Exec(func(handle Handle) error {
			err := handle.CreateSchema([]interface{}{&models.User{}, &models.Bucket{}})
			if common.Error(err) {
				return err
			}

			return nil
		})

		if common.Error(err) {
			return err
		}
	}

	return nil
}

func createModel(modelPath string, modelFileInfo os.FileInfo) error {
	type data struct {
		Type string
		Name string
	}

	var d data

	if modelFileInfo.IsDir() || filepath.Base(modelPath) == "base.go" || filepath.Base(modelPath) == "class.go" {
		return nil
	}

	databasePath := common.CleanPath(filepath.Join(filepath.Dir(modelPath), "..", "service", "database"))

	model := common.FileNamePart(filepath.Base(modelPath))

	for _, typ := range []string{"mongo", "pgsql"} {
		outputFile := filepath.Join(databasePath, fmt.Sprintf("%s_%s.go", typ, model))

		if !common.FileExists(outputFile) {
			return &common.ErrFileNotFound{
				FileName: outputFile,
			}
		}

		b, err := os.ReadFile(filepath.Join(databasePath, fmt.Sprintf("%s_class.go", typ)))
		if err != nil {
			return err
		}

		t := template.Must(template.New(".").Parse(string(b)))

		var buf bytes.Buffer

		err = t.Execute(&buf, d)
		if err != nil {
			return err
		}

		code := string(buf.Bytes())

		for i := 5; i > 1; i-- {
			code = strings.Replace(code, "    ", "\t", -1)
		}

		searchReplaces := []struct {
			search  string
			reaplce string
		}{
			{
				search:  "Class",
				reaplce: common.Capitalize(model),
			},
			{
				search:  "class",
				reaplce: model,
			},
		}

		for _, v := range searchReplaces {
			search := v.search
			replace := v.reaplce

			code = strings.Replace(code, search, replace, -1)
		}

		err = os.WriteFile(common.CleanPath(outputFile), []byte(code), common.DefaultFileMode)
		if common.Error(err) {
			return err
		}
	}

	return nil

}

func Codegen() error {
	fw, err := common.NewFilewalker(filepath.Join("models", "*.go"), false, false, createModel)

	if common.Error(err) {
		return err
	}

	err = fw.Run()
	if common.Error(err) {
		return err
	}

	return nil
}

func Close() {
	if pool == nil {
		return
	}

	close(pool)
	for handle := range pool {
		common.Error(handle.Stop())
	}

	common.Info("Service database stopped")
}

func Get() Handle {
	handle := <-pool

	return handle
}

func Put(handle Handle) {
	pool <- handle
}

func Exec(fn func(handle Handle) error) error {
	handle := Get()
	defer Put(handle)

	return fn(handle)
}

func create(cfg *Cfg) (Handle, error) {
	var handle Handle
	var err error

	switch cfg.Driver {
	case TYPE_MONGODB:
		handle, err = NewMongoDB()
	case TYPE_PGSQL:
		handle, err = NewPgsqlDB()
	default:
		return nil, &errors.ErrUnknownDriver{Driver: cfg.Driver}
	}

	if common.Error(err) {
		return nil, err
	}

	err = handle.Init(cfg)
	if common.Error(err) {
		return nil, err
	}

	err = handle.Start()
	if common.Error(err) {
		return nil, err
	}

	return handle, nil
}
