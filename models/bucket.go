package models

//go:generate templater -sr Class=Bucket;class=document -i ../service/database/storm_class.go -o ../service/database/storm_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/badger_class.go -o ../service/database/badger_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/mongo_class.go -o ../service/database/mongo_bucket.go
//go:generate templater -sr Class=Bucket;class=document -i ../service/database/pgsql_class.go -o ../service/database/pgsql_bucket.go

type Bucket struct {
	Base     `storm:"inline"`
	Uid      string            `storm:"unique"`
	Prop     map[string]string `sql:",hstore"`
	FileName []string          `sql:",array"`
	FileType []string          `sql:",array"`
	FileLen  []int64           `sql:",array"`
	FileHash []string          `sql:",array"`
}
