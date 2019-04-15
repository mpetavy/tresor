package database

import (
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *StormDB) SaveClass(class *models.Class, options *Options) error {
	if class.CreatedAt.IsZero() {
		class.CreatedAt = time.Now()
	} else {
		class.ModifiedAt = time.Now()
	}

	return db.DB.Save(class)
}

func (db *StormDB) LoadClass(field string, value interface{}, class *models.Class, options *Options) error {
	return db.DB.One(field, value, class)
}

func (db *StormDB) DeleteClass(field string, value interface{}, id int, options *Options) error {
	class := &models.Class{}

	err := db.LoadClass(field, value, class, options)
	if err != nil {
		return err
	}

	return db.DB.DeleteStruct(class)
}
