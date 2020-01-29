package database

import (
	"github.com/mpetavy/tresor/models"
	"time"
)

func (db *PgsqlDB) SaveBucket(document *models.Bucket, options *Options) error {
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	} else {
		document.ModifiedAt = time.Now()
	}

	return db.ORM.Insert(document)
}

func (db *PgsqlDB) LoadBucket(field string, value interface{}, document *models.Bucket, options *Options) error {
	return nil
}

func (db *PgsqlDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	return nil
}
