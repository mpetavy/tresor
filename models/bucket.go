package models

import (
	"context"
	"github.com/go-pg/pg/orm"
	"time"
)

//go:generate templater -sr Class=Bucket;class=document -i ../service/database/storm_class.go -o ../service/database/storm_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/badger_class.go -o ../service/database/badger_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/mongo_class.go -o ../service/database/mongo_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/pgsql_class.go -o ../service/database/pgsql_bucket.go

type Bucket struct {
	Base     `storm:"inline"`
	Uid      string            `sql:",unique" storm:",unique"`
	Prop     map[string]string `sql:",hstore" sqlindex:"gin"`
	FileName []string          `sql:",array" sqlindex:"gin"`
	FileType []string          `sql:",array" sqlindex:"gin"`
	FileLen  []int64           `sql:",array"`
	FileHash []string          `sql:",array"`
}

func NewBucket() Bucket {
	b := Bucket{}
	b.Prop = make(map[string]string)

	return b
}

func (b *Bucket) BeforeInsert(c context.Context, db orm.DB) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}

	return nil
}
