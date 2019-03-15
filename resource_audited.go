package admin

import (
	"github.com/moisespsena-go/aorm"
)

func (res *Resource) configureAudited() {
	if _, ok := res.Value.(aorm.Auditor); ok {
		for _, fname := range aorm.AuditedFields {
			if m := res.Meta(&Meta{Name: fname}); m.Enabled == nil {
				res.Meta(&Meta{
					Name: fname,
					Type: "-",
				})
			}
		}
	}

	if res.softDelete {
		for _, fname := range aorm.SoftDeleteFields {
			res.Meta(&Meta{
				Name: fname,
				Type: "-",
			})
		}
	}
}
