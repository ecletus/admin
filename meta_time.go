package admin

import (
	"reflect"
	"strings"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
)

type TimeConfig struct {
	TimeFormat           string
	TimeFormatFunc       func(ctx *core.Context) string
	LocationFallback     *time.Location
	LocationFallbackFunc func(ctx *core.Context) *time.Location
	locationFalbackValue string
}

func (this *TimeConfig) LocationFalbackValue() string {
	return this.locationFalbackValue
}

func (this *TimeConfig) FormatC(ctx *core.Context) (v string) {
	if this.TimeFormat != "" {
		return this.TimeFormat
	}
	if this.TimeFormatFunc != nil {
		return this.TimeFormatFunc(ctx)
	}
	return ctx.Ts(I18NGROUP+".metas.time.format", "hh:mm")
}

func (this *TimeConfig) Layout(ctx *core.Context) (layout string) {
	format := this.FormatC(ctx)
	var part string
	var done = func() {
		switch part {
		case "h", "hh":
			layout += "15"
		case "m", "mm":
			layout += "04"
		case "SS", "S":
			layout += "05"
		case "s":
			layout += "000"
		case "ss":
			layout += "000000"
		case "sss":
			layout += "000000000"
		case "z":
			layout += "-0700"
		case "Z":
			layout += "-07:00"
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

func (this *TimeConfig) FormattedValue(meta *Meta, value interface{}, context *core.Context) interface{} {
	var t time.Time
	switch date := meta.GetValuer()(value, context).(type) {
	case *time.Time:
		if date == nil {
			return ""
		}
		if date.IsZero() {
			return ""
		}
		t = *date
	case time.Time:
		if date.IsZero() {
			return ""
		}
		t = date
	default:
		return date
	}

	if t.IsZero() {
		return ""
	}
	var loc *time.Location
	if meta.Resource != nil {
		loc = OptGetLocation(meta.Resource, ContextFromCoreContext(context))
	} else {
		loc = OptGetLocation(meta.BaseResource, ContextFromCoreContext(context))
	}
	if loc == nil {
		loc = context.TimeLocation
	}
	t = t.In(loc)
	return utils.FormatTime(t, this.FormatC(context), context)
}

// ConfigureQorMeta configure select one meta
func (this *TimeConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "time"
	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) interface{} {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil {
		TimeConfigSetter(meta)
	}
}

func (this *TimeConfig) Configure() {
	if this.LocationFallback == nil {
		this.LocationFallback = time.Local
	}
	this.locationFalbackValue = time.Now().In(this.LocationFallback).Format("-0700")
}

func (this *TimeConfig) LocationFallbackC(ctx *core.Context) *time.Location {
	if this.LocationFallbackFunc == nil {
		return this.LocationFallback
	}
	return this.LocationFallbackFunc(ctx)
}

func (this *TimeConfig) LocationFallbackValue(ctx *core.Context) string {
	if this.LocationFallbackFunc == nil {
		return this.locationFalbackValue
	}

	return time.Now().In(this.LocationFallbackFunc(ctx)).Format("-0700")
}

func (this *TimeConfig) Parse(ctx *core.Context, value ...string) ([]time.Time, error) {
	return ParseTimeArgs(this.Layout(ctx), this.LocationFallbackValue(ctx), value...)
}

func ParseTimeArgs(layout, fallbackLoc string, values ...string) (times []time.Time, err error) {
	var loc string
	if fallbackLoc != "" {
		hasLoc := strings.Contains(layout, "07")
		if !hasLoc {
			layout += " -0700"
		}
		if !hasLoc {
			loc = " " + fallbackLoc
		}
	}
	times = make([]time.Time, len(values))
	for i, value := range values {
		if value == "" {
			continue
		}
		value += loc
		var t time.Time
		if t, err = time.Parse(layout, value); err != nil {
			return nil, err
		}
		times[i] = t
	}
	return
}

func TimeConfigSetter(meta *Meta) {
	cfg := meta.Config.(interface {
		Parse(ctx *core.Context, value ...string) ([]time.Time, error)
	})
	meta.Meta.Setter = resource.SingleFieldSetter(meta.Meta, meta.FieldName, func(_ bool, field reflect.Value, metaValue *resource.MetaValue, context *core.Context, record interface{}) (err error) {
		var times []time.Time
		if times, err = cfg.Parse(context, metaValue.FirstStringValue()); err != nil {
			return err
		}
		utils.SetNonZero(field, times[0])
		return nil
	})
}
