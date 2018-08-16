package admin

import (
	"fmt"
	"strings"
	"github.com/aghape/aghape"
	"github.com/moisespsena-go/aorm"
)

func resourceParents(res *Resource) []*Resource {
	var parents []*Resource
	r := res.ParentResource
	for i := 0; r != nil; i++ {
		parents = append(parents, r)
		r = r.ParentResource
	}
	return parents
}

func resourceParamName(parents []*Resource, param string) string {
	var names []string
	for i := len(parents); i > 0; i-- {
		names = append(names, parents[i-1].Param)
	}
	names = append(names, param)
	return strings.Join(names, "__")
}

func resourceParamIDName(level int, paramName string) string {
	return fmt.Sprintf("resource_%02d__%v__id", level, paramName)
}

func subResourceConfigureFilters(res *Resource) {
	res.DefaultFilter(func(context *qor.Context, db *aorm.DB) *aorm.DB {
		return res.FilterByParent(db, context.URLParam(res.ParentResource.paramIDName))
	})

	res.DefaultFilter(res.Config.Sub.Filters...)

	scope := res.FakeScope

	if res.Config.Sub.RawFieldFilter != nil {
		var rawDbFields []string
		var rawDbFieldsValues []interface{}
		for fieldName, value := range res.Config.Sub.RawFieldFilter {
			if f, ok := scope.FieldByName(fieldName); ok {
				rawDbFields = append(rawDbFields, scope.QuotedTableName() + "." + f.DBName)
				rawDbFieldsValues = append(rawDbFieldsValues, value)
			} else {
				panic("Field \"" + fieldName + "\" does not exists.")
			}
		}
		res.DefaultFilter(func(context *qor.Context, db *aorm.DB) *aorm.DB {
			return db.Where(strings.Join(rawDbFields, " AND "), rawDbFieldsValues...)
		})
	}
}