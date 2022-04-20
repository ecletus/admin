package admin

import "github.com/go-aorm/aorm"

type dbKey uint8

const DbKey dbKey = 1

func FromDb(db *aorm.DB) *Admin {
	if v, ok := db.Get(DbKey); ok {
		return v.(*Admin)
	}
	return nil
}
