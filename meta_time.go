package admin

import (
	"bytes"
	"reflect"
	"strings"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

const (
	TimeConfigDateTime = iota + 1
	TimeConfigDate
	TimeConfigTime
)

type TimeConfigType uint8

func (t TimeConfigType) FallbackLayouts() []string {
	switch t {
	case TimeConfigDateTime:
		return dateTimeLayouts
	case TimeConfigDate:
		return dateLayouts
	case TimeConfigTime:
		return timeLayouts
	default:
		return nil
	}
}

type TimeConfig struct {
	TimeFormat           string
	TimeFormatFunc       func(ctx *core.Context) string
	LocationFallback     *time.Location
	LocationFallbackFunc func(ctx *core.Context) *time.Location
	locationFalbackValue string
	LocationConfigor     core.Configor
	I18nKey              string
	DefaultFormat        string
	Type                 TimeConfigType
	TimeType             aorm.TimeType
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
	return ctx.Ts(this.I18nKey, this.DefaultFormat)
}

func (this *TimeConfig) Layout(ctx *core.Context) (layout string) {
	return ParseTimeLayout(this.FormatC(ctx))
}

func (this *TimeConfig) FormattedValue(meta *Meta, record interface{}, context *core.Context) *FormattedValue {
	var t time.Time
	switch vt := record.(type) {
	case time.Time:
		t = vt
	case *time.Time:
		if vt != nil {
			t = *vt
		}
	default:
		switch date := meta.Value(context, record).(type) {
		case *time.Time:
			if date == nil {
				return nil
			}
			if date.IsZero() {
				return nil
			}
			t = *date
		case time.Time:
			if date.IsZero() {
				return nil
			}
			t = date
		default:
			panic("bad time value")
		}
	}

	if t.IsZero() {
		return &FormattedValue{Record: record}
	}
	return (&FormattedValue{Record: record, Raw: t, Value: this.Format(context, t)}).SetNonZero()
}

func (this *TimeConfig) Format(context *core.Context, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return utils.FormatTime(t, this.Layout(context), context)
}

// ConfigureQorMeta configure select one meta
func (this *TimeConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	this.Setup()
	if meta.Type == "" {
		meta.Type = "time"
	}
	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) *FormattedValue {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil {
		meta.Meta.Setter = resource.SingleFieldSetter(meta.FieldName, func(_ bool, field reflect.Value, metaValue *resource.MetaValue, context *core.Context, record interface{}) (err error) {
			var times []time.Time
			if times, err = this.Parse(context, metaValue.StringValue()); err != nil {
				return err
			}
			utils.SetNonZero(field, times[0])
			return nil
		})
	}
}

func (this *TimeConfig) GetFilterValue(arg *FilterArgument) (val interface{}, err error) {
	if arg.Value != nil {
		var (
			times   []time.Time
			v       = []string{arg.Value.GetString("Value")}
			noBlank bool
		)
		if v[0] == "" {
			v = []string{arg.Value.GetString("Start"), arg.Value.GetString("End")}
		}
		for _, v := range v {
			if noBlank = v != ""; noBlank {
				break
			}
		}
		if noBlank {
			if times, err = ParseTimeArgs(this.Layout(arg.Context), this.Type.FallbackLayouts(), this.LocationFallbackValue(arg.Context),
				v...); err != nil {
				return
			}
			switch len(times) {
			case 1:
				return aorm.NewTimeValue(times[0], this.TimeType), nil
			case 2:
				return TimeRange{
					aorm.NewTimeValue(times[0], this.TimeType),
					aorm.NewTimeValue(times[1], this.TimeType),
				}, nil
			}
		}
	}
	return
}

func (this *TimeConfig) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Type == "" {
		filter.Type = "time"
	}
	if filter.Valuer == nil {
		filter.Valuer = this.GetFilterValue
	}
	this.Setup()
}

func (this *TimeConfig) Setup() {
	if this.I18nKey == "" {
		this.I18nKey = I18NGROUP + ".metas.time.format"
	}

	if this.DefaultFormat == "" {
		this.DefaultFormat = "hh:mm"
	}

	if this.LocationFallback != nil {
		this.locationFalbackValue = time.Now().In(this.LocationFallback).Format("-0700")
	}
	if this.Type == 0 {
		this.Type = TimeConfigTime
	}
}

func (this *TimeConfig) LocationFallbackC(ctx *core.Context) *time.Location {
	if this.LocationFallbackFunc == nil {
		if this.LocationFallback == nil {
			return OptGetLocation(this.LocationConfigor, ContextFromCoreContext(ctx))
		}
		return this.LocationFallback
	}
	return this.LocationFallbackFunc(ctx)
}

func (this *TimeConfig) LocationFallbackValue(ctx *core.Context) string {
	if this.LocationFallbackFunc == nil {
		if this.locationFalbackValue == "" {
			t := time.Now().In(this.LocationFallbackC(ctx))
			return t.Format("-0700")
		}
		return this.locationFalbackValue
	}
	return time.Now().In(this.LocationFallbackFunc(ctx)).Format("-0700")
}

func (this *TimeConfig) Parse(ctx *core.Context, value ...string) ([]time.Time, error) {
	return ParseTimeArgs(this.Layout(ctx), this.Type.FallbackLayouts(), this.LocationFallbackValue(ctx), value...)
}

func ParseTimeArgs(layout string, fallbackLayouts []string, fallbackLoc string, values ...string) (times []time.Time, err error) {
	var (
		loc         string
		layoutSufix string
	)
	if fallbackLoc != "" {
		hasLoc := strings.Contains(layout, "07")
		if !hasLoc {
			layoutSufix = " -0700"
		}
		if !hasLoc {
			loc = " " + fallbackLoc
		}
	}

	times = make([]time.Time, len(values))

values_loop:
	for i, value := range values {
		if value == "" {
			continue
		}
		value += loc
		var t time.Time
		if t, err = time.Parse(layout+layoutSufix, value); err == nil {
			times[i] = t
			continue
		}
		var err2 error
		for _, l2 := range fallbackLayouts {
			if t, err2 = time.Parse(l2, value); err2 == nil {
				times[i] = t
				err = nil
				continue values_loop
			}
			l2 = l2 + layoutSufix
			if t, err2 = time.Parse(l2, value); err2 == nil {
				times[i] = t
				err = nil
				continue values_loop
			}
		}
		return nil, err
	}
	return
}

var (
	dateLayouts = []string{
		"2006-01-02",
	}

	dateTimeLayouts = []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}

	timeLayouts = []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
)

var ParseTimeLayout = func(format string) string {
	var (
		part string
		buf  bytes.Buffer
	)

	var done = func() {
		switch part {
		case "h", "hh":
			buf.WriteString("15")
		case "mm":
			buf.WriteString("04")
		case "m":
			buf.WriteString("4")
		case "ss":
			buf.WriteString("05")
		case "s":
			buf.WriteString("5")
		case "S":
			buf.WriteString("000")
		case "SS":
			buf.WriteString("000000")
		case "SSS":
			buf.WriteString("000000000")
		case "z":
			buf.WriteString("-0700")
		case "Z":
			buf.WriteString("-07:00")
		case "dd", "d":
			buf.WriteString("02")
		case "MM", "M":
			buf.WriteString("01")
		case "yyyy", "y":
			buf.WriteString("2006")
		case "yy":
			buf.WriteString("06")
		default:
			buf.WriteString(part)
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
	return buf.String()
}

type TimeValue = aorm.TimeValue
