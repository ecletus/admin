package admin

import (
	"github.com/pkg/errors"
)

func SingleEdit(r *Resource, field ...string) {
	Admin := r.GetAdmin()
	m := func(fieldName string) {
		id_ := r.FullID() + "." + fieldName
		if err := Admin.OnResourcesAdded(func(e *ResourceEvent) error {
			r.SetMeta(&Meta{Name: fieldName, Type: "single_edit", Resource: e.Resource})
			return nil
		}, id_); err != nil {
			panic(errors.Wrap(err, id_))
		}
	}

	for _, name := range field {
		m(name)
	}
}

func SingleEditPairs(r *Resource, fieldName_resource_pairs ...interface{}) {
	for i, l := 0, len(fieldName_resource_pairs); i < l; i += 2 {
		fieldName, res := fieldName_resource_pairs[i].(string), fieldName_resource_pairs[i+1].(*Resource)
		r.SetMeta(&Meta{Name: fieldName, Type: "single_edit", Resource: res})
	}
}
