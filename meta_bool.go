package admin

import (
	"reflect"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

type SelectOneBoolConfig struct {
	Truth, Falsy         string
	TruthFunc, FalsyFunc func(ctx *Context, record interface{}) string
	meta                 *Meta
}

func (this *SelectOneBoolConfig) ConfigureQorMeta(metaor resource.Metaor) {
	m := metaor.(*Meta)
	if this.meta != nil && this.meta != m {
		copy := *this
		*this = copy
	}
	m.Type = "select_one_bool"
	this.meta = m
	if m.Typ.Kind() == reflect.Ptr {
		m.SetFormattedValuer(func(recorde interface{}, ctx *core.Context) interface{} {
			value := m.Value(ctx, recorde)
			if value == nil {
				return ""
			}
			b := value.(*bool)
			if b == nil {
				return nil
			}
			if *b {
				return this.TruthLabel(ContextFromContext(ctx), recorde)
			}
			return this.FalsyLabel(ContextFromContext(ctx), recorde)
		})
	} else {
		m.NewValuer(func(meta *Meta, old MetaValuer, recorde interface{}, ctx *core.Context) interface{} {
			if m.BaseResource.GetKey(recorde).IsZero() {
				return nil
			}
			if value := old(recorde, ctx); m.IsZero(recorde, value) {
				return nil
			} else {
				return value
			}
		})
		m.NewSetter(func(meta *Meta, old MetaSetter, recorde interface{}, metaValue *resource.MetaValue, ctx *core.Context) error {
			if err := old(recorde, metaValue, ctx); err != nil {
				return err
			}
			metaValue.NoBlank = true
			return nil
		})
		m.SetFormattedValuer(func(recorde interface{}, ctx *core.Context) interface{} {
			value := m.Value(ctx, recorde)
			if m.IsZero(recorde, value) {
				return nil
			}
			if value.(bool) {
				return this.TruthLabel(ContextFromContext(ctx), recorde)
			}
			return this.FalsyLabel(ContextFromContext(ctx), recorde)
		})
	}
}

func (this *SelectOneBoolConfig) TruthLabel(ctx *Context, record interface{}) string {
	if this.TruthFunc != nil {
		return this.TruthFunc(ctx, record)
	}
	if this.Truth == "" {
		return ctx.Ts(I18NGROUP+".form.bool.true", "Yes")
	}
	return ctx.Ts(this.Truth, "Yes")
}

func (this *SelectOneBoolConfig) FalsyLabel(ctx *Context, record interface{}) string {
	if this.FalsyFunc != nil {
		return this.FalsyFunc(ctx, record)
	}
	if this.Falsy == "" {
		return ctx.Ts(I18NGROUP+".form.bool.false", "No")
	}
	return ctx.Ts(this.Falsy, "Yes")
}

func init() {
	RegisterMetaTypeConfigor(reflect.TypeOf(true), func(meta *Meta) {
		if meta.Config == nil {
			if meta.IsRequired() || meta.Typ.Kind() == reflect.Ptr {
				cfg := &SelectOneBoolConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)
			} else {
				meta.Type = "switch"
			}
		}
	})
}
