package admin

import (
	"github.com/moisespsena-go/aorm"
)

func (this *Resource) configureAudited() {
	if _, ok := this.Value.(aorm.Auditor); ok {
		for _, fname := range aorm.AuditedFields {
			this.Meta(&Meta{Name: fname, Label: "aorm.audited.fields." + fname, DefaultInvisible: true})
		}
	}

	if this.softDelete {
		for _, fname := range aorm.SoftDeleteFields {
			this.Meta(&Meta{Name: fname, Label: "aorm.soft_delete.fields." + fname, DefaultInvisible: true})
		}
	}
}
