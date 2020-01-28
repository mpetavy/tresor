package database

import (
	"github.com/mpetavy/common"
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *StormDB) SaveUser(user *models.User, options *Options) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	} else {
		user.ModifiedAt = time.Now()
	}

	return db.DB.Save(user)
}

func (db *StormDB) LoadUser(field string, value interface{}, user *models.User, options *Options) error {
	return db.DB.One(field, value, user)
}

func (db *StormDB) DeleteUser(field string, value interface{}, id int, options *Options) error {
	user := &models.User{}

	err := db.LoadUser(field, value, user, options)
	if common.Error(err) {
		return err
	}

	return db.DB.DeleteStruct(user)
}
