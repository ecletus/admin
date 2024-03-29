package admin

import (
	"fmt"
	"reflect"

	"github.com/go-aorm/aorm"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/jinzhu/copier"
)

func indirectType(r reflect.Type) reflect.Type {
	for r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	return r
}
func IndirectRealType(r reflect.Type) reflect.Type {
	for r.Kind() == reflect.Ptr || r.Kind() == reflect.Interface {
		r = r.Elem()
	}
	return r
}
func indirectValuePtr(v reflect.Value) *reflect.Value {
	v = reflect.Indirect(v)
	return &v
}

type ProxyPath interface {
	GetType() reflect.Type
	Get(ctx *core.Context, recorde *reflect.Value) (value *reflect.Value)
}

type proxyPathField struct {
	Index []int
	Type  reflect.Type
}

func (p proxyPathField) Get(ctx *core.Context, recorde *reflect.Value) (value *reflect.Value) {
	v := recorde.FieldByIndex(p.Index)
	return &v
}

func (p proxyPathField) GetType() reflect.Type {
	return p.Type
}

type proxyPathVirtualField struct {
	Name string
	Type reflect.Type
}

func (p proxyPathVirtualField) GetType() reflect.Type {
	return p.Type
}

func (p proxyPathVirtualField) Get(ctx *core.Context, recorde *reflect.Value) (value *reflect.Value) {
	recordeVF := recorde.Addr().Interface().(aorm.VirtualFieldsGetter)
	result, ok := recordeVF.GetVirtualField(p.Name)
	if !ok {
		return nil
	}
	v := reflect.ValueOf(result)
	return &v
}

type proxyPathMeta struct {
	Meta *Meta
}

func (p proxyPathMeta) GetType() reflect.Type {
	panic("not implemented!")
}

func (p proxyPathMeta) Get(ctx *core.Context, recorde *reflect.Value) (value *reflect.Value) {
	metaRecorde := p.Meta.Value(ctx, recorde.Addr().Interface())
	if metaRecorde == nil {
		return nil
	}
	v := reflect.ValueOf(metaRecorde)
	return &v
}

type proxyPathGetter struct {
	Getter func(ctx *core.Context, recorde interface{}) (value interface{})
	Type   reflect.Type
}

func (p proxyPathGetter) GetType() reflect.Type {
	return p.Type
}

func (p proxyPathGetter) Get(ctx *core.Context, recorde *reflect.Value) (value *reflect.Value) {
	vi := p.Getter(ctx, recorde.Addr().Interface())
	if vi == nil {
		return nil
	}
	v := reflect.ValueOf(vi)
	return &v
}

type ProxyVirtualFieldPath struct {
	FieldName string
	Value     interface{}
}

type ProxyMetaPath struct {
	Meta *Meta
}

type ProxyPathGetter struct {
	Get   func(ctx *core.Context, recorde interface{}) interface{}
	Value interface{}
}

func NewMetaFieldProxy(name string, parts []interface{}, src interface{}, to *Meta) *Meta {
	var (
		path []ProxyPath
		ro   bool
	)

	r := reflect.TypeOf(src)
	for _, p := range parts {
		r = indirectType(r)
		switch pt := p.(type) {
		case string:
			if f, ok := r.FieldByName(pt); ok {
				path = append(path, proxyPathField{f.Index, f.Type})
				r = f.Type
			} else {
				panic(fmt.Errorf("Invalid path"))
			}
		case ProxyVirtualFieldPath:
			path = append(path, proxyPathVirtualField{pt.FieldName, reflect.TypeOf(pt.Value)})
		case ProxyMetaPath:
			path = append(path, proxyPathMeta{pt.Meta})
			if !ro && pt.Meta.ReadOnly {
				ro = true
			}
		case ProxyPathGetter:
			path = append(path, proxyPathGetter{pt.Get, reflect.TypeOf(pt.Value)})
		default:
			panic(fmt.Errorf("Invalid path"))
		}
	}

	meta := NewMetaProxy(name, to, nil)
	meta.proxyPath = path
	meta.ReadOnly = ro
	return meta
}

func NewMetaProxy(name string, to *Meta, recorde func(meta *Meta, recorde interface{}) interface{}) *Meta {
	meta := &Meta{}
	copier.Copy(meta, to)
	meta.ProxyTo = to
	to.Proxies = append(to.Proxies, meta)
	meta.Meta = &resource.Meta{}
	copier.Copy(meta.Meta, to.Meta)
	to.AfterUpdate(func() {
		meta.Config = to.Config
	})
	meta.GetRecordHandler = func(ctx *core.Context, r interface{}) (rec interface{}) {
		if r == nil {
			return
		}
		if meta.proxyPath != nil {
			value := indirectValuePtr(reflect.ValueOf(r))
			for _, pth := range meta.proxyPath {
				if value = pth.Get(ctx, value); value == nil {
					return nil
				} else if !value.IsValid() {
					return nil
				} else if value.Kind() == reflect.Ptr {
					if value.IsNil() {
						return nil
					}
					value = indirectValuePtr(*value)
				}
			}
			if value.Kind() == reflect.Struct {
				if !value.CanAddr() {
					v2 := reflect.New(value.Type())
					v2.Elem().Set(*value)
					return v2.Interface()
				}
				return value.Addr().Interface()
			}
			return value.Interface()
		}
		return recorde(meta, r)
	}

	meta.Name = name
	if to.Setter != nil {
		meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *core.Context) error {
			return to.Setter(resource, metaValue, context)
		}
	}
	if to.ContextResourcer != nil {
		meta.ContextResourcer = func(meta resource.Metaor, context *core.Context) resource.Resourcer {
			return to.ContextResourcer(meta, context)
		}
	}
	if to.GetMetasFunc != nil {
		meta.GetMetasFunc = func() []resource.Metaor {
			return to.GetMetasFunc()
		}
	}
	meta.ForceShowZero = to.ForceShowZero
	return meta
}
