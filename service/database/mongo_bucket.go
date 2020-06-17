package database

import (
	"context"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (db *MongoDB) SaveBucket(bucket *models.Bucket, options *Options) error {
	if bucket.CreatedAt.IsZero() {
		bucket.CreatedAt = time.Now()
	} else {
		bucket.ModifiedAt = time.Now()
	}

	b, err := bson.Marshal(bucket)
	if common.Error(err) {
		return err
	}

	collection := db.Client.Database(db.Name).Collection("bucket")
	_, err = collection.InsertOne(context.Background(), b)

	return err
}

func (db *MongoDB) LoadBucket(field string, value interface{}, bucket *models.Bucket, options *Options) error {
	return nil
}

func (db *MongoDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	return nil
}
