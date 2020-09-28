package admin

import (
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/edis"
)

const (
	E_FOREIGN_META_ADDED = "foreignMetaAdded"
)

func (this *Resource) triggerForeignMetaAdded(meta *Meta) {
	if err := this.Trigger(&ForeignMetaEvent{edis.NewEvent(E_FOREIGN_META_ADDED), this, meta}); err != nil {
		panic(err)
	}
	if err := this.Trigger(&ForeignMetaEvent{edis.NewEvent(E_FOREIGN_META_ADDED + ":" + meta.ID()), this, meta}); err != nil {
		panic(err)
	}
}

func (this *Resource) OnForeignMetaAdded(cb func(meta *Meta)) {
	this.On(E_FOREIGN_META_ADDED, func(e edis.EventInterface) {
		cb(e.(*ForeignMetaEvent).Meta)
	})
	for _, meta := range this.ForeignMetas {
		cb(meta)
	}
}

func MetaTypeAdded(typ interface{}, cb func(meta *Meta)) func(meta *Meta) {
	rtype := utils.IndirectType(typ)
	return func(meta *Meta) {
		if indirectType(meta.Typ) == rtype {
			cb(meta)
		}
	}
}
