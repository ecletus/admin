package admin

import (
	"fmt"
	"regexp"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/xroute"
)

// TODO: unused

var primaryKeyRegexp = regexp.MustCompile(`primary_key\[.+_.+\]`)

func (this *Admin) registerCompositePrimaryKeyCallback(router xroute.Router) {
	router.Use(&xroute.Middleware{
		Name: PKG + ".composite_primary_key_filter",
		Handler: func(chain *xroute.ChainHandler) {
			context := core.ContextFromRequest(chain.Request()).Value(CONTEXT_KEY).(*Context)
			for key, value := range context.Request.URL.Query() {
				if primaryKeyRegexp.MatchString(key) {
					context.DB(context.DB().Set(key, value))
				}
			}
			chain.Pass()
		},
	})
}

var DisableCompositePrimaryKeyMode = PKG + ".composite_primary_key:query:disable"

func compositePrimaryKeyQueryCallback(scope *aorm.Scope) {
	if scope.Value == nil {
		return
	}
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
