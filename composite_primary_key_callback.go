package admin

import (
	"fmt"
	"regexp"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

var primaryKeyRegexp = regexp.MustCompile(`primary_key\[.+_.+\]`)

func (admin *Admin) registerCompositePrimaryKeyCallback() {
	// register middleware
	router := admin.GetRouter()
	router.Use(&Middleware{
		Name: "composite primary key filter",
		Handler: func(context *Context, middleware *Middleware) {
			db := context.DB
			for key, value := range context.Request.URL.Query() {
				if primaryKeyRegexp.MatchString(key) {
					db = db.Set(key, value)
				}
			}
			context.DB = db

			middleware.Next(context)
		},
	})

	admin.Config.SetupDB(func(db *qor.DB) error {
		if db.DB.Callback().Query().Get("qor_admin:composite_primary_key") == nil {
			db.DB.Callback().Query().Before("gorm:query").Register("qor_admin:composite_primary_key", compositePrimaryKeyQueryCallback)
		}

		if db.DB.Callback().RowQuery().Get("qor_admin:composite_primary_key") == nil {
			db.DB.Callback().RowQuery().Before("gorm:row_query").Register("qor_admin:composite_primary_key", compositePrimaryKeyQueryCallback)
		}

		return nil
	})
}

var DisableCompositePrimaryKeyMode = "composite_primary_key:query:disable"

func compositePrimaryKeyQueryCallback(scope *gorm.Scope) {
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
