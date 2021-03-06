package models

//go:generate templater -sr Class=User;class=user -i ../service/database/mongo_class.go -o ../service/database/mongo_user.go
//go:generate templater -sr Class=User;class=user -i ../service/database/pgsql_class.go -o ../service/database/pgsql_user.go

type User struct {
	Base     `storm:"inline"`
	Name     string `storm:"unique"`
	Password string
}

func NewUser() User {
	u := User{}

	return u
}
