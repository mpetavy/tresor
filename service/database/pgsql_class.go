package database

import (
	"github.com/mpetavy/tresor/models"
	"time"
)

func (db *PgsqlDB) SaveClass(class *models.Class, options *Options) error {
	if class.CreatedAt.IsZero() {
		class.CreatedAt = time.Now()
	} else {
		class.ModifiedAt = time.Now()
	}

	return db.DB.Insert(class)
}

func (db *PgsqlDB) LoadClass(field string, value interface{}, class *models.Class, options *Options) error {
	return nil
}

func (db *PgsqlDB) DeleteClass(field string, value interface{}, id int, options *Options) error {
	return nil
}
