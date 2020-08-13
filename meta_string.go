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
		if t, ok := meta.Tags.Get("MAX_LEN"); ok {
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
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) interface{} {
			switch str := meta.Value(context, value).(type) {
			case *string:
				if str != nil {
					return *str
				}
				return ""
			case string:
				return str
			default:
				return str
			}
		})
	}
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &StringConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		} else {
			meta.Config.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("string", cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(""), cfg)
}
