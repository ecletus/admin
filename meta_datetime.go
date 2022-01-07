package admin

import (
	"reflect"
	"time"

	"github.com/ecletus/core/resource"
)

type DateTimeConfig struct {
	TimeConfig
}

func (this *DateTimeConfig) Setup() {
	if this.I18nKey == "" {
		this.I18nKey = I18NGROUP + ".metas.datetime.format"
	}
	if this.DefaultFormat == "" {
		this.DefaultFormat = "yyyy-MM-dd hh:mm"
	}
	this.Type = TimeConfigDateTime
}

// ConfigureQorMeta configure select one meta
func (this *DateTimeConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	if meta.Type == "" {
		meta.Type = "datetime"
	}
	this.Setup()
	this.TimeConfig.ConfigureQorMeta(metaor)
}

func (this *DateTimeConfig) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Type == "" {
		filter.Type = "datetime"
	}
	this.Setup()
	this.TimeConfig.ConfigureQORAdminFilter(filter)
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
