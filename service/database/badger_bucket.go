package database

import (
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *BadgerDB) SaveBucket(document *models.Bucket, options *Options) error {
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	} else {
		document.ModifiedAt = time.Now()
	}

	return nil
}

func (db *BadgerDB) LoadBucket(field string, value interface{}, document *models.Bucket, options *Options) error {
	return nil
}

func (db *BadgerDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	return nil
}
