package admin

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ecletus/core"
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core/resource"
)

type UintConfig struct {
	Step     uint8
	Max, Min uint64
	Setter   func(recorde interface{}, value uint64)
}

func (this *UintConfig) HtmlAttributtes() template.HTML {
	var (
		attrs = []string{}
		step  = this.Step
	)
	if step == 0 {
		step = 1
	}
	attrs = append(attrs, `step="`+fmt.Sprint(this.Step)+`"`)
	if this.Max > 0 {
		attrs = append(attrs, `max="`+fmt.Sprint(this.Max)+` "`)
	}
	attrs = append(attrs, `min="`+fmt.Sprint(this.Min)+` "`)
	return template.HTML(strings.Join(attrs, " "))
}

// ConfigureQorMeta configure select one meta
func (this *UintConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "uint"

	if meta.Tags.Flag("%") {
		this.Max = 100

		if meta.DefaultFormat == "" {
			meta.DefaultFormat = "%d%%"
		}
	}

	if s := meta.Tags.Get("MIN"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(err)
		}
		if this.Min == 0 || v > this.Min {
			this.Min = v
		}
	}

	if s := meta.Tags.Get("MAX"); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(err)
		}
		if this.Max == 0 || v < this.Max {
			this.Max = v
		}
	}

	if this.Min > 0 && this.Max > 0 && this.Min > this.Max {
		panic(fmt.Errorf("UintConfig: meta %s of %s: min > max", meta.BaseResource.UID, meta.Name, this.Min, this.Max))
	}

	if meta.Setter == nil {
		meta.Meta.Setter = resource.SingleFieldSetter(meta.FieldName, func(ptr bool, field reflect.Value, metaValue *resource.MetaValue, ctx *core.Context, record interface{}) (err error) {
			var (
				v = metaValue.FirstStringValue()
				i uint64
			)
			if v != "" {
				if i, err = strconv.ParseUint(v, 10, 64); err != nil {
					return
				}
			}
			if this.Setter != nil {
				this.Setter(record, i)
			} else if ptr {
				field.Elem().SetUint(i)
			} else {
				field.SetUint(i)
			}
			return
		})
	}
}

func init() {
	cfg := func(meta *Meta) {
		if meta.Config == nil {
			cfg := &UintConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}
	RegisterMetaConfigor("uint", cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(uint64(0)), cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(uint32(0)), cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(uint16(0)), cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(uint8(0)), cfg)
	RegisterMetaTypeConfigor(reflect.TypeOf(uint(0)), cfg)
}
