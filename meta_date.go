package admin

import (
	"strings"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"

	"github.com/ecletus/core/resource"
)

type DateConfig struct {
	DateFormat     string
	DateFormatFunc func(ctx *core.Context) string
}

func (this *DateConfig) FormatC(ctx *core.Context) (v string) {
	defer func() {
		v = strings.ToLower(v)
	}()
	if this.DateFormat != "" {
		return this.DateFormat
	}
	if this.DateFormatFunc != nil {
		return this.DateFormatFunc(ctx)
	}
	return ctx.Ts(I18NGROUP+".metas.date.format", "yyyy-mm-dd")
}

func (this *DateConfig) Layout(ctx *core.Context) (layout string) {
	format := this.FormatC(ctx)
	var part string
	var done = func() {
		switch part {
		case "dd", "d":
			layout += "02"
		case "mm", "m":
			layout += "01"
		case "yyyy", "y":
			layout += "2006"
		case "HH", "H":
			layout += "15"
		case "MM", "M":
			layout += "04"
		case "SS", "S":
			layout += "05"
		case "s":
			layout += "000"
		case "ss":
			layout += "000000"
		case "sss":
			layout += "000000000"
		default:
			layout += part
		}
	}
	for _, c := range format {
		if len(part) == 0 || c == rune(part[0]) {
			part += string(c)
		} else {
			done()
			part = string(c)
		}
	}
	done()
	return
}

func (this *DateConfig) FormattedValue(meta *Meta, record interface{}, context *core.Context) interface{} {
	switch date := meta.Value(context, record).(type) {
	case *time.Time:
		if date == nil {
			return ""
		}
		if date.IsZero() {
			return ""
		}
		return utils.FormatTime(*date, this.Layout(context), context)
	case time.Time:
		if date.IsZero() {
			return ""
		}
		return utils.FormatTime(date, this.Layout(context), context)
	default:
		return date
	}
}

// ConfigureQorMeta configure select one meta
func (this *DateConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "date"
	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) interface{} {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil {
		TimeConfigSetter(meta)
	}
}


func (this *DateConfig) Parse(ctx *core.Context, value ...string) ([]time.Time, error) {
	return ParseTimeArgs(this.Layout(ctx), "", value...)
}