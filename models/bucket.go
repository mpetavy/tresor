package models

import (
	"context"
	"github.com/go-pg/pg/orm"
	"time"
)

//go:generate templater -sr Class=Bucket;class=bucket -i ../service/database/mongo_class.go -o ../service/database/mongo_bucket.go
//go:generate templater -sr Class=Bucket;class=bucket -i ../service/database/pgsql_class.go -o ../service/database/pgsql_bucket.go

type Bucket struct {
	Base            `storm:"inline"`
	Uid             string            `sql:",unique" storm:",unique"`
	Props           map[string]string `sql:",hstore" sqlx:"gin"`
	FileNames       []string          `sql:",array" sqlx:"gin"`
	FileMimeTypes   []string          `sql:",array" sqlx:"gin"`
	FileSizes       []int64           `sql:",array"`
	FileHashes      []string          `sql:",array"`
	FileFulltext    []string          `sql:",array"`
	FileOrientation []int             `sql:",array"`
}

func NewBucket() Bucket {
	b := Bucket{}
	b.Props = make(map[string]string)

	return b
}

func (b *Bucket) BeforeInsert(c context.Context, db orm.DB) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}

	return nil
}
