package admin_helpers

import (
	"reflect"

	"github.com/aghape/admin"
	"github.com/aghape/core/utils"
)

func SingleEdit(r *admin.Resource, field ...string) {
	typ := utils.IndirectType(r.Value)

	Admin := r.GetAdmin()
	m := func(fieldName string) {
		field, _ := typ.FieldByName(fieldName)
		value := reflect.New(utils.IndirectType(field.Type)).Interface()
		_ = Admin.OnResourceValueAdded(value, func(e *admin.ResourceEvent) {
			r.SetMeta(&admin.Meta{Name: fieldName, Type: "single_edit", Resource: e.Resource})
		})
	}

	for _, name := range field {
		m(name)
	}
}

func SingleEditPairs(r *admin.Resource, fieldName_resource_pairs ...interface{}) {
	for i, l := 0, len(fieldName_resource_pairs); i < l; i += 2 {
		fieldName, res := fieldName_resource_pairs[i].(string), fieldName_resource_pairs[i+1].(*admin.Resource)
		r.SetMeta(&admin.Meta{Name: fieldName, Type: "single_edit", Resource: res})
	}
}
