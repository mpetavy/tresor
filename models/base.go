package models

import (
	"time"
)

type Base struct {
	Id         int       `sql:",pk" storm:"id,increment"`
	CreatedAt  time.Time `sql:",notnull,default:now()"`
	ModifiedAt time.Time `sql:",notnull,default:now()"`
}
