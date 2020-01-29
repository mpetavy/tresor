package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	_ "github.com/lib/pq"
	"github.com/mpetavy/common"
	"reflect"
	"strings"
)

type PgsqlDB struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	ORM      *pg.DB
	DB       *sql.DB
}

func NewPgsqlDB() (*PgsqlDB, error) {
	return &PgsqlDB{}, nil
}

func (db *PgsqlDB) Init(cfg *common.Jason) error {
	var err error

	db.Username, err = cfg.String("username")
	if common.Error(err) {
		return err
	}
	db.Password, err = cfg.String("password")
	if common.Error(err) {
		return err
	}
	db.Host, err = cfg.String("host")
	if common.Error(err) {
		return err
	}
	db.Port, err = cfg.Int("port")
	if common.Error(err) {
		return err
	}
	db.Database, err = cfg.String("database")
	if common.Error(err) {
		return err
	}

	connStr := fmt.Sprintf("user='%s' password='%s' host='%s' port='%d' dbname='%s' sslmode='disable'",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database)

	db.DB, err = sql.Open("postgres", connStr)
	if common.Error(err) {
		return err
	}

	return nil
}

func (db *PgsqlDB) CreateSchema(models []interface{}) error {
	for _, model := range models {
		err := db.ORM.DropTable(model, &orm.DropTableOptions{})
		if common.Error(err) {
			common.Warn(err.Error())
		}

		err = db.ORM.CreateTable(model, &orm.CreateTableOptions{})
		if common.Error(err) {
			return err
		}

		for _, f := range structs.Fields(model) {
			tag := f.Tag("sqlx")

			if tag != "" {
				tableName := structs.Name(model) + "s"
				indexName := tableName + "__" + strings.ToLower(f.Name())

				indexType := "("
				if tag == "gin" || f.Kind() == reflect.Map || f.Kind() == reflect.Array {
					indexType = "using gin ("
				}
				indexType += underscore(f.Name())
				indexType += ")"

				q := fmt.Sprintf("create index %s on %s %s", indexName, tableName, indexType)

				_, err = db.ORM.Exec(q)
				if common.Error(err) {
					return err
				}
			}
		}
	}

	return nil
}

func (db *PgsqlDB) SwitchIndices(models []interface{}, enable bool) error {
	for _, model := range models {
		tableName := structs.Name(model) + "s"

		q := fmt.Sprintf("update pg_index set indisready=%v where indrelid=(select oid from pg_class where relname='%s')", enable, tableName)
		_, err := db.ORM.Exec(q)
		if common.Error(err) {
			return err
		}

		if enable {
			q := fmt.Sprintf("reindex table %s", tableName)
			_, err := db.ORM.Exec(q)
			if common.Error(err) {
				return err
			}
		}
	}

	return nil
}

func (db *PgsqlDB) SQL(query string) (string, error) {
	var objects []map[string]interface{}

	rows, err := db.DB.Query(query)
	if err != nil {
		return "", err
	}
	defer func() {
		common.Error(rows.Close())
	}()

	for rows.Next() {
		columns, err := rows.ColumnTypes()
		if err != nil {
			return "", err
		}

		values := make([]interface{}, len(columns))
		object := map[string]interface{}{}
		for i, column := range columns {
			object[column.Name()] = reflect.New(column.ScanType()).Interface()
			values[i] = object[column.Name()]
		}

		err = rows.Scan(values...)
		if err != nil {
			return "", err
		}

		objects = append(objects, object)
	}

	ba, err := json.MarshalIndent(objects, "", "    ")
	if common.Error(err) {
		return "", err
	}

	return string(ba), nil
}

func (db *PgsqlDB) Start() error {
	db.ORM = pg.Connect(&pg.Options{
		User:     db.Username,
		Password: db.Password,
		Addr:     fmt.Sprintf("%s:%d", db.Host, db.Port),
		Database: db.Database,
	})

	return nil
}

func (db *PgsqlDB) Stop() error {
	if db.ORM != nil {
		common.Error(db.ORM.Close())
	}

	if db.DB != nil {
		common.Error(db.DB.Close())
	}

	return nil
}

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func toUpper(c byte) byte {
	return c - 32
}

func toLower(c byte) byte {
	return c + 32
}

func underscore(s string) string {
	r := make([]byte, 0, len(s)+5)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUpper(c) {
			if i > 0 && i+1 < len(s) && (isLower(s[i-1]) || isLower(s[i+1])) {
				r = append(r, '_', toLower(c))
			} else {
				r = append(r, toLower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}
