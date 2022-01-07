package admin

import (
	"fmt"
	"reflect"

	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type UrlConfig struct {
	Target string
	Download,
	Copy, // Copy to clipboard
	NoLink bool
	Label                   string
	LabelField, LabelMethod string

	ReadonlyLabelEnabled bool
	LabelFunc            func(ctx *Context, record interface{}) string
	WrapFunc             func(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML
	meta                 *Meta
}

func (this *UrlConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if this.meta == nil {
		meta := metaor.(*Meta)
		this.meta = meta
		meta.Type = "url"
		if this.Label == "" && this.LabelFunc == nil {
			if tags := meta.Tags.GetTags("CONFIG"); tags != nil {
				if this.Label = tags.Get("LABEL"); this.Label == "" {
					if fieldName := tags.Get("LABEL_FIELD"); fieldName != "" {
						this.LabelField = fieldName
					} else if methodName := tags.Get("LABEL_METHOD"); methodName != "" {
						this.LabelMethod = methodName
					}
				}
			}

			if this.LabelField != "" {
				if field, ok := this.meta.BaseResource.ModelStruct.Type.FieldByName(this.LabelField); ok {
					this.LabelFunc = func(ctx *Context, record interface{}) string {
						return reflect.Indirect(reflect.ValueOf(record)).FieldByIndex(field.Index).Interface().(string)
					}
				} else {
					panic(fmt.Errorf("MetaUrl: Field %q for %q does not exists", this.LabelField, this.meta.BaseResource.ModelStruct.Fqn()))
				}
			} else if this.LabelMethod != "" {
				if m, ok := reflect.PtrTo(this.meta.BaseResource.ModelStruct.Type).MethodByName(this.LabelMethod); ok {
					if m.Type.NumIn() == 1 {
						this.LabelFunc = func(ctx *Context, record interface{}) string {
							res := reflect.ValueOf(record).Method(m.Index).Call([]reflect.Value{})
							return res[0].Interface().(string)
						}
					} else {
						this.LabelFunc = func(ctx *Context, record interface{}) string {
							res := reflect.ValueOf(record).Method(m.Index).Call([]reflect.Value{reflect.ValueOf(ctx)})
							return res[0].Interface().(string)
						}
					}
				} else {
					panic(fmt.Errorf("MetaUrl: Method %q for %q does not exists", this.LabelMethod, this.meta.BaseResource.ModelStruct.Fqn()))
				}
			}
		}
	}
}

func (this *UrlConfig) GetLabel(ctx *Context, record interface{}) string {
	if this.LabelFunc != nil {
		return this.LabelFunc(ctx, record)
	}
	return this.Label
}

func (this *UrlConfig) Wrap(s *template.State, ctx *Context, record interface{}, value template.HTML) template.HTML {
	if this.WrapFunc != nil {
		return this.WrapFunc(s, ctx, record, value)
	}
	return value
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &UrlConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("url", cfg)
}
