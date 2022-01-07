package admin

import (
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) configureAudited() {
	for _, name := range append(aorm.AuditedFields, "CreatedBy", "UpdatedBy") {
		if _, ok := this.ModelStruct.FieldsByName[name]; ok {
			this.Meta(&Meta{Name: name, Label: "aorm.audited.fields." + name, DefaultInvisible: true})
		}
	}

	for _, name := range append(aorm.SoftDeleteFields, "DeletedBy") {
		if _, ok := this.ModelStruct.FieldsByName[name]; ok {
			this.Meta(&Meta{Name: name, Label: "aorm.audited.fields." + name, DefaultInvisible: true})
		}
	}

	for _, name := range append(aorm.SoftDeletionDisableFields, "DeletionDisabledBy") {
		if _, ok := this.ModelStruct.FieldsByName[name]; ok {
			this.Meta(&Meta{Name: name, Label: "aorm.audited.fields." + name, DefaultInvisible: true})
		}
	}
}
