package admin

import (
	"fmt"
	"strings"

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
	res.DefaultFilter(&DBFilter{
		Name: "admin:parent_filter",
		Handler: func(context *Context, db *aorm.DB) (DB *aorm.DB, err error) {
			if context.ResourceID == nil && len(context.ParentResourceID) > 0 {
				if parentId := context.ParentResourceID[len(context.ParentResourceID)-1]; parentId != nil {
					if res.Config.Sub.ParentFilter != nil {
						return res.Config.Sub.ParentFilter(context.Context, db, parentId)
					}
					return res.FilterByParent(context.Context, db, parentId)
				}
			}
			return db, nil
		},
	})

	res.DefaultFilters.AddFilter(res.Config.Sub.Filters...)

	if res.Config.Sub.RawFieldFilter != nil {
		var rawDbFields []string
		var rawDbFieldsValues []interface{}
		for fieldName, value := range res.Config.Sub.RawFieldFilter {
			if f, ok := res.ModelStruct.FieldsByName[fieldName]; ok {
				rawDbFields = append(rawDbFields, "{}."+f.DBName)
				rawDbFieldsValues = append(rawDbFieldsValues, value)
			} else {
				panic("Field \"" + fieldName + "\" does not exists.")
			}
		}
		res.DefaultFilter(&DBFilter{
			Name: "admin:parent_filter:raw_fields",
			Handler: func(context *Context, db *aorm.DB) (DB *aorm.DB, err error) {
				return db.Where(aorm.IQ(strings.Join(rawDbFields, " AND ")), rawDbFieldsValues...), nil
			},
		})
	}
}
