package admin

import (
	"fmt"
	"reflect"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
)

type MaskConfig struct {
	MaskFunc func(context *core.Context, record interface{}) string
	Mask     string
	Unmask   func(context *core.Context, record interface{}, maskedValue string) string
	JsCode   string
	JsCoder  func(context *Context, record interface{}) string
}

func (this *MaskConfig) GetJsCode(context *Context, record interface{}) string {
	if this.JsCoder != nil {
		return this.JsCoder(context, record)
	}
	if this.JsCode != "" {
		return this.JsCode
	}

	var mask = this.Mask
	if mask == "" && this.MaskFunc != nil {
		mask = this.MaskFunc(context.Context, record)
	}
	return `this.mask("` + mask + `")`
}

// ConfigureQorMeta configure meta
func (this *MaskConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "string_mask"
	if meta.Setter == nil && this.Unmask != nil {
		meta.Meta.Setter = resource.SingleFieldSetter(meta.Meta, meta.FieldName, func(_ bool, field reflect.Value, metaValue *resource.MetaValue, context *core.Context, record interface{}) (err error) {
			if value := metaValue.FirstStringValue(); value == "" {
				value = this.Unmask(context, record, value)
				utils.SetNonZero(field, value)
			} else {
				utils.SetZero(field)
			}
			return nil
		})
	}
}

func init() {
	RegisterMetaConfigor("string_mask", func(meta *Meta) {
		if meta.Config == nil {
			cfg := &MaskConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	})

	RegisterMetaConfigureTagsHandler(func(meta *Meta, tags *MetaTags) {
		if meta.Config != nil {
			return
		}
		if maskTags := tags.Tags("MASK"); maskTags != nil {
			var code = fmt.Sprintf("this.mask(%q", maskTags["code"])
			if maskTags.Flag("reverse") {
				code += ", {reverse: true}"
			} else if options := maskTags["options"]; options != "" {
				code += ", "+options
			}
			code += ")"
			m := &MaskConfig{JsCode: code}
			meta.Config = m
			m.ConfigureQorMeta(meta)
		} else if mask := tags.TagSetting["MASK"]; mask != "" {
			if m := GetMask(mask); m != nil {
				meta.Config = m
				m.ConfigureQorMeta(meta)
			} else {
				m = &MaskConfig{JsCode: mask}
				meta.Config = m
				m.ConfigureQorMeta(meta)
			}
		}
	})
}
