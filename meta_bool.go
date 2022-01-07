package admin

import (
	"reflect"
	"strings"

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

	var label = func(v bool, ctx *core.Context, record interface{}) string {
		if v {
			return this.TruthLabel(ContextFromContext(ctx), record)
		}
		return this.FalsyLabel(ContextFromContext(ctx), record)
	}

	if m.Typ.Kind() == reflect.Ptr {
		m.SetFormattedValuer(func(recorde interface{}, ctx *core.Context) *FormattedValue {
			value := m.Value(ctx, recorde)
			if value == nil {
				return (&FormattedValue{Record: recorde}).SetZero()
			}
			b := value.(*bool)
			if b == nil {
				return nil
			}
			return (&FormattedValue{Record: recorde, Raw: value, Value: label(*b, ctx, recorde)}).SetNonZero()
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
		m.SetFormattedValuer(func(record interface{}, ctx *core.Context) *FormattedValue {
			value := m.Value(ctx, record)
			if !value.(bool) {
				return nil
			}
			return (&FormattedValue{Record: record, Raw: value, Value: label(value.(bool), ctx, record)}).SetNonZero()
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
	if strings.ContainsRune(this.Truth, ':') {
		return ctx.Ts(this.Truth, "Yes")
	}
	return this.Truth
}

func (this *SelectOneBoolConfig) FalsyLabel(ctx *Context, record interface{}) string {
	if this.FalsyFunc != nil {
		return this.FalsyFunc(ctx, record)
	}
	if this.Falsy == "" {
		return ctx.Ts(I18NGROUP+".form.bool.false", "No")
	}
	if strings.ContainsRune(this.Falsy, ':') {
		return ctx.Ts(this.Falsy, "No")
	}
	return this.Falsy
}

func init() {
	RegisterMetaTypeConfigor(reflect.TypeOf(true), func(meta *Meta) {
		if meta.Config == nil {
			if meta.IsRequired() || meta.Typ.Kind() == reflect.Ptr {
				cfg := &SelectOneBoolConfig{}
				meta.Config = cfg

				if values := meta.Tags.GetTags("BOOL"); values != nil {
					cfg.Truth, cfg.Falsy = values.Get("T"), values.Get("F")
				}

				cfg.ConfigureQorMeta(meta)
			} else {
				meta.Type = "switch"
			}
		}
	})
}
