package database

import (
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *BadgerDB) SaveClass(class *models.Class, options *Options) error {
	if class.CreatedAt.IsZero() {
		class.CreatedAt = time.Now()
	} else {
		class.ModifiedAt = time.Now()
	}

	return nil
}

func (db *BadgerDB) LoadClass(field string, value interface{}, class *models.Class, options *Options) error {
	return nil
}

func (db *BadgerDB) DeleteClass(field string, value interface{}, id int, options *Options) error {
	return nil
}
