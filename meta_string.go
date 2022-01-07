package admin

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
)

type StringConfig struct {
	MaxLen uint16
}

func (this *StringConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	if meta.Type == "" {
		meta.Type = "string"
	}
	if this.MaxLen == 0 {
		if t, ok := meta.Tags.GetOk("MAX_LEN"); ok {
			ui64, err := strconv.ParseUint(t, 10, 16)
			if err != nil {
				panic(err)
			}
			this.MaxLen = uint16(ui64)
		} else if meta.FieldStruct != nil {
			this.MaxLen = uint16(meta.FieldStruct.TextSize())
		}
	}

	if this.MaxLen > 0 {
		meta.Validator(func(record interface{}, values *resource.MetaValue, ctx *core.Context) (err error) {
			if v := values.FirstStringValue(); v != "" && len(v) > int(this.MaxLen) {
				msg := fmt.Sprintf(ctx.Ts(I18NGROUP+".errors.validations.too_long_text"), this.MaxLen)
				return resource.ErrField(ctx, record, values.Meta.GetFieldName(), values.Meta.GetRecordLabelC(ctx, record))(msg)
			}
			return
		})
	}

	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(record interface{}, context *core.Context) *FormattedValue {
			value := meta.Value(context, record)
			if value == nil {
				return nil
			}
			switch str := value.(type) {
			case *string:
				if str != nil {
					return (&FormattedValue{Record: record, Raw: str, Value: *str}).SetNonZero()
				}
				return nil
			case string:
				if str == "" {
					return nil
				}
				return (&FormattedValue{Record: record, Raw: str, Value: str}).SetNonZero()
			default:
				fv := (&FormattedValue{Record: record, Raw: value}).SetNonZero()
				if meta.Tags.Flag("SAFE") {
					fv.SafeValue = ContextFromCoreContext(context).Stringify(value)
				}
				return fv
			}
		})
	}
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			if meta.Type != "password" && meta.FieldStruct != nil {
				var tags = meta.FieldStruct.TagSettings
				if size, ok := tags["SIZE"]; ok {
					if i, _ := strconv.Atoi(size); i > 255 {
						meta.Type = "text"
					} else {
						meta.Type = "string"
					}
				} else if text, ok := tags["TYPE"]; ok && text == "text" {
					meta.Type = "text"
				} else {
					meta.Type = "string"
				}
			}

			if meta.Type == "text" {
				cfg := &TextConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)
			} else {
				cfg := &StringConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)
			}
		} else {
			meta.Config.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("string", cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(""), cfg)
}
