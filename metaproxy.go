package admin

import (
	"reflect"

	"github.com/jinzhu/copier"
	"github.com/aghape/aghape"
	"github.com/aghape/aghape/resource"
	"github.com/aghape/aghape/serializer"
)

func NewMetaFieldProxy(name string, parts []string, to *Meta) *Meta {
	return NewMetaProxy(name, to, func(meta *Meta, i interface{}) interface{} {
		if i == nil {
			return nil
		}

		r := reflect.Indirect(reflect.ValueOf(i))
		for _, p := range parts[0 : len(parts)-2] {
			r = r.FieldByName(p)
			if !r.IsValid() && r.Type().Implements(serializer.SerializableFieldType) {
				if ri, ok := r.Interface().(serializer.SerializableField).GetSerializableField(p); ok {
					r = reflect.ValueOf(ri)
				}
			} else {
				r = reflect.Indirect(r)
			}
		}

		r = r.FieldByName(parts[len(parts)-2])
		if !r.IsValid() || r.IsNil() {
			return nil
		}
		return r.Interface()
	})
}

func NewMetaProxy(name string, to *Meta, recorde func(meta *Meta, recorde interface{}) interface{}) *Meta {
	meta := &Meta{}
	copier.Copy(meta, to)
	meta.ProxyTo = to
	meta.Meta = &resource.Meta{}
	copier.Copy(meta.Meta, to.Meta)
	record := func(r interface{}) interface{} {
		return recorde(meta, r)
	}

	meta.Name = name
	if to.Valuer != nil {
		meta.Valuer = func(i interface{}, context *qor.Context) interface{} {
			return to.Valuer(record(i), context)
		}
	}
	if to.FormattedValuer != nil {
		meta.FormattedValuer = func(i interface{}, context *qor.Context) interface{} {
			return to.FormattedValuer(record(i), context)
		}
	}
	if to.TypeHandler != nil {
		meta.TypeHandler = func(i interface{}, context *Context, meta *Meta) string {
			return to.TypeHandler(record(i), context, meta)
		}
	}
	if to.Enabled != nil {
		meta.Enabled = func(i interface{}, context *Context, meta *Meta) bool {
			return to.Enabled(record(i), context, meta)
		}
	}
	if to.Setter != nil {
		meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) error {
			return to.Setter(resource, metaValue, context)
		}
	}
	if to.ContextResourcer != nil {
		meta.ContextResourcer = func(meta resource.Metaor, context *qor.Context) resource.Resourcer {
			return to.ContextResourcer(meta, context)
		}
	}
	if to.GetMetasFunc != nil {
		meta.GetMetasFunc = func() []resource.Metaor {
			return to.GetMetasFunc()
		}
	}
	if to.ShowRenderIgnore != nil {
		meta.ShowRenderIgnore = func(recorde, value interface{}) bool {
			return to.ShowRenderIgnore(record(recorde), value)
		}
	}
	meta.ForceShowRender = to.ForceShowRender
	return meta
}
