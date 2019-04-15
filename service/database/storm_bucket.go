package database

import (
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *StormDB) SaveBucket(document *models.Bucket, options *Options) error {
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	} else {
		document.ModifiedAt = time.Now()
	}

	return db.DB.Save(document)
}

func (db *StormDB) LoadBucket(field string, value interface{}, document *models.Bucket, options *Options) error {
	return db.DB.One(field, value, document)
}

func (db *StormDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	document := &models.Bucket{}

	err := db.LoadBucket(field, value, document, options)
	if err != nil {
		return err
	}

	return db.DB.DeleteStruct(document)
}
