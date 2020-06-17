package database

import (
	"github.com/mpetavy/tresor/models"
	"time"
)

func (db *PgsqlDB) SaveBucket(bucket *models.Bucket, options *Options) error {
	if bucket.CreatedAt.IsZero() {
		bucket.CreatedAt = time.Now()
	} else {
		bucket.ModifiedAt = time.Now()
	}

	return db.ORM.Insert(bucket)
}

func (db *PgsqlDB) LoadBucket(field string, value interface{}, bucket *models.Bucket, options *Options) error {
	return nil
}

func (db *PgsqlDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	return nil
}
