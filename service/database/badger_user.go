package database

import (
	"time"

	"github.com/mpetavy/tresor/models"
)

func (db *BadgerDB) SaveUser(user *models.User, options *Options) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	} else {
		user.ModifiedAt = time.Now()
	}

	return nil
}

func (db *BadgerDB) LoadUser(field string, value interface{}, user *models.User, options *Options) error {
	return nil
}

func (db *BadgerDB) DeleteUser(field string, value interface{}, id int, options *Options) error {
	return nil
}
