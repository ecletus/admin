package admin

import (
	"reflect"
	"time"

	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"
)

type DateTimeConfig struct {
	DateConfig
	TimeConfig
}

func (this *DateTimeConfig) FormattedValue(meta *Meta, value interface{}, context *core.Context) interface{} {
	var t time.Time
	switch date := meta.GetValuer()(value, context).(type) {
	case *time.Time:
		if date != nil {
			t = *date
		}
	case time.Time:
		t = date
	default:
		return date
	}
	if !t.IsZero() {
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
		return utils.FormatTime(t, this.Layout(context), context)
	}
	return ""
}

// ConfigureQorMeta configure select one meta
func (this *DateTimeConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	meta.Type = "datetime"
	if meta.FormattedValuer == nil {
		meta.SetFormattedValuer(func(value interface{}, context *core.Context) interface{} {
			return this.FormattedValue(meta, value, context)
		})
	}
	if meta.Setter == nil || resource.IsDefaultMetaSetter(meta.Meta) {
		TimeConfigSetter(meta)
	}
}

func (this *DateTimeConfig) Format(ctx *core.Context) string {
	return this.DateConfig.FormatC(ctx) + " " + this.TimeConfig.FormatC(ctx)
}

func (this *DateTimeConfig) Layout(ctx *core.Context) string {
	return this.DateConfig.Layout(ctx) + " " + this.TimeConfig.Layout(ctx)
}

func (this *DateTimeConfig) Parse(ctx *core.Context, value ...string) ([]time.Time, error) {
	return ParseTimeArgs(this.Layout(ctx), this.LocationFallbackValue(ctx), value...)
}

func init() {
	metaConfigorMaps["datetime"] = func(meta *Meta) {
		if meta.Config == nil {
			cfg := &DateTimeConfig{}
			meta.Config = cfg
			cfg.ConfigureQorMeta(meta)
		}
	}

	metaTypeConfigorMaps[reflect.TypeOf(time.Time{})] = func(meta *Meta) {
		if meta.Config != nil || meta.Type != "" {
			return
		}

		if meta.FieldStruct != nil {
			typ := meta.FieldStruct.TagSettings["TYPE"]
			switch typ {
			case "date":
				cfg := &DateConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)

			case "time", "timetz", "timz":
				cfg := &TimeConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)

			case "", "datetime", "datetimetz", "datetimez":
				cfg := &DateTimeConfig{}
				meta.Config = cfg
				cfg.ConfigureQorMeta(meta)
			}
		}
	}
}