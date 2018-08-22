package admincommon

import (
	"fmt"
	"time"
	"reflect"
	"strings"
	
	"gopkg.in/oleiade/reflections.v1"
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/admin"
	"github.com/aghape/core"
)

func RecordInfoFields(r *admin.Resource) {
	reflectValue := reflect.TypeOf(r.Value)

	hasField := func(fieldName string) bool {
		_, ok := reflectValue.FieldByName(fieldName)
		return ok
	}

	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	if _, found := reflectValue.FieldByName("DeletedAt"); found {
		r.Scope(&admin.Scope{Name: "All", Default: true,
			Handler: func(db *aorm.DB, s *admin.Searcher, context *core.Context) *aorm.DB {
				if db.NewScope(r.Value).Search.Unscoped || s.CurrentScopes.Has("trash") {
					return db
				}

				return db.Where(db.NewScope(r.Value).QuotedTableName() + ".deleted_at IS NULL")
			}})

		r.Scope(&admin.Scope{Name: "trash", Handler: func(db *aorm.DB, s *admin.Searcher, context *core.Context) *aorm.DB {
			return db.Where(db.NewScope(r.Value).QuotedTableName() + ".deleted_at IS NOT NULL").Unscoped()
		}})

		restorable := func(record interface{}) bool {
			v, err := reflections.GetField(record, "DeletedAt")
			if err == nil && v != nil {
				if t, ok := v.(*time.Time); ok {
					if t != nil {
						if ! t.IsZero() {
							return true
						}
					}
				}
			}
			return false
		}

		r.Action(&admin.Action{
			Name:  "Restore",
			Modes: []string{"index", "menu_item"},
			Visible: func(record interface{}, context *admin.Context) bool {
				return restorable(record)
			},
			Handler: func(argument *admin.ActionArgument) error {
				argument.Context.SetDB(argument.Context.DB.Unscoped())
				scope := argument.Context.DB.NewScope(r.Value)

				for _, record := range argument.FindSelectedRecords() {
					if restorable(record) {
						err := argument.Context.DB.Model(record).Updates(map[string]interface{}{"DeletedAt": nil, "DeletedBy": nil}).Error
						if err != nil {
							var idv []string
							for _, pf := range scope.PrimaryFields() {
								v, _ := reflections.GetField(record, pf.Name)
								idv = append(idv, fmt.Sprint(pf.Name, ": ", v))
							}
							return fmt.Errorf(strings.Join(idv, " & "), " restore error: ", err)
						}
					}
				}
				return nil
			},
		})
	}

	for _, fname := range ([]string{"CreatedById", "CreatedAt", "DeletedById", "DeletedAt", "UpdatedById", "UpdatedAt"}) {
		if hasField(fname) {
			r.Meta(&admin.Meta{Name: fname, Type: "-"})
		}
	}
}
