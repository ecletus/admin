package admin

import (
	"fmt"
	"regexp"

	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-route"
)

var primaryKeyRegexp = regexp.MustCompile(`primary_key\[.+_.+\]`)

func (admin *Admin) registerCompositePrimaryKeyCallback() {
	// register middleware
	router := admin.Router
	router.Use(&route.Middleware{
		Name: PKG + ".composite_primary_key_filter",
		Handler: func(chain *route.ChainHandler) {
			context := ContextFromChain(chain)
			db := context.DB
			for key, value := range context.Request.URL.Query() {
				if primaryKeyRegexp.MatchString(key) {
					db = db.Set(key, value)
				}
			}
			context.DB = db
			chain.Pass()
		},
	})
}

var DisableCompositePrimaryKeyMode = PKG + ".composite_primary_key:query:disable"

func compositePrimaryKeyQueryCallback(scope *aorm.Scope) {
	if value, ok := scope.Get(DisableCompositePrimaryKeyMode); ok && value != "" {
		return
	}

	tableName := scope.TableName()
	for _, primaryField := range scope.PrimaryFields() {
		if value, ok := scope.Get(fmt.Sprintf("primary_key[%v_%v]", tableName, primaryField.DBName)); ok && value != "" {
			scope.Search.Where(fmt.Sprintf("%v = ?", scope.Quote(primaryField.DBName)), value)
		}
	}
}
