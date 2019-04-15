package models

import (
	"time"
)

type Base struct {
	Id         int `storm:"id,increment"`
	CreatedAt  time.Time
	ModifiedAt time.Time
}
