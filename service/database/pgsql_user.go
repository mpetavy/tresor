package database

import (
	"github.com/mpetavy/tresor/models"
	"time"
)

func (db *PgsqlDB) SaveUser(user *models.User, options *Options) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	} else {
		user.ModifiedAt = time.Now()
	}

	return db.ORM.Insert(user)
}

func (db *PgsqlDB) LoadUser(field string, value interface{}, user *models.User, options *Options) error {
	return nil
}

func (db *PgsqlDB) DeleteUser(field string, value interface{}, id int, options *Options) error {
	return nil
}
