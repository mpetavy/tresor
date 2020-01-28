package database

import (
	"context"
	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (db *MongoDB) SaveClass(class *models.Class, options *Options) error {
	if class.CreatedAt.IsZero() {
		class.CreatedAt = time.Now()
	} else {
		class.ModifiedAt = time.Now()
	}

	b, err := bson.Marshal(class)
	if common.Error(err) {
		return err
	}

	collection := db.Client.Database(db.Name).Collection("class")
	_, err = collection.InsertOne(context.Background(), b)

	return err
}

func (db *MongoDB) LoadClass(field string, value interface{}, class *models.Class, options *Options) error {
	return nil
}

func (db *MongoDB) DeleteClass(field string, value interface{}, id int, options *Options) error {
	return nil
}
