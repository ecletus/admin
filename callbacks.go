package admin

import (
	"reflect"

	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/core/utils"
)

func (this *Admin) OnResourceTypeAdded(typ interface{}, cb func(res *Resource)) {
	rtyp := utils.StructType(typ)
	if this.onResourceTypeAdded == nil {
		this.onResourceTypeAdded = map[reflect.Type][]func(res *Resource){}
	}
	if _, ok := this.onResourceTypeAdded[rtyp]; ok {
		this.onResourceTypeAdded[rtyp] = append(this.onResourceTypeAdded[rtyp], cb)
	} else {
		this.onResourceTypeAdded[rtyp] = []func(res *Resource){}
	}
	if resources, ok := this.ResourcesByType[rtyp]; ok {
		for _, res := range resources {
			cb(res)
		}
	}
}

func (this *Admin) OnPreInitializeMeta(f ...func(meta *Meta)) {
	this.onPreInitializeResourceMeta = append(this.onPreInitializeResourceMeta, f...)
}

func (this *Admin) OnRouter(f ...func(r xroute.Router)) {
	this.onRouter = append(this.onRouter, f...)
}

func (this *Admin) AddNewContextCallback(callback func(ctx *Context)) *Admin {
	this.NewContextCallbacks = append(this.NewContextCallbacks, callback)
	return this
}

// RegisterMetaConfigor register configor for a kind, it will be called when register those kind of metas
func (this *Admin) RegisterMetaConfigor(kind string, fc func(*Meta)) {
	this.metaConfigorMaps[kind] = fc
}

// RegisterFuncMap register view funcs, it could be used in view templates
func (this *Admin) RegisterFuncMap(name string, fc interface{}) {
	this.funcMaps[name] = fc
}
