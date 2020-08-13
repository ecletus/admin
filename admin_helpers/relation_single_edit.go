package admin_helpers

import (
	"github.com/ecletus/admin"
	"github.com/pkg/errors"
)

func SingleEdit(r *admin.Resource, field ...string) {
	Admin := r.GetAdmin()
	m := func(fieldName string) {
		id_ := r.FullID() + "." + fieldName
		if err := Admin.OnResourcesAdded(func(e *admin.ResourceEvent) error {
			r.SetMeta(&admin.Meta{Name: fieldName, Type: "single_edit", Resource: e.Resource})
			return nil
		}, id_); err != nil {
			panic(errors.Wrap(err, id_))
		}
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
