package database

import (
	"context"
	"github.com/mpetavy/tresor/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (db *MongoDB) SaveUser(user *models.User, options *Options) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	} else {
		user.ModifiedAt = time.Now()
	}

	b, err := bson.Marshal(user)
	if err != nil {
		return err
	}

	collection := db.Client.Database(db.Name).Collection("user")
	_, err = collection.InsertOne(context.Background(), b)

	return err
}

func (db *MongoDB) LoadUser(field string, value interface{}, user *models.User, options *Options) error {
	return nil
}

func (db *MongoDB) DeleteUser(field string, value interface{}, id int, options *Options) error {
	return nil
}
