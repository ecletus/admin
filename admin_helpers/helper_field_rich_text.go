package admin_helpers

import "github.com/ecletus/admin"

func FieldRichEditor(r *admin.Resource, field ...string) {
	for _, fieldName := range field {
		r.Meta(&admin.Meta{Name: fieldName, Config: &admin.RichEditorConfig{}})
	}
}
