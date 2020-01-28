package database

import (
	"context"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (db *MongoDB) SaveBucket(document *models.Bucket, options *Options) error {
	if document.CreatedAt.IsZero() {
		document.CreatedAt = time.Now()
	} else {
		document.ModifiedAt = time.Now()
	}

	b, err := bson.Marshal(document)
	if common.Error(err) {
		return err
	}

	collection := db.Client.Database(db.Name).Collection("document")
	_, err = collection.InsertOne(context.Background(), b)

	return err
}

func (db *MongoDB) LoadBucket(field string, value interface{}, document *models.Bucket, options *Options) error {
	return nil
}

func (db *MongoDB) DeleteBucket(field string, value interface{}, id int, options *Options) error {
	return nil
}
