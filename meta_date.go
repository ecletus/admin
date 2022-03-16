package admin

import (
	"github.com/ecletus/core/resource"
	"github.com/moisespsena-go/aorm"
)

type DateConfig struct {
	TimeConfig
}

func NewDateConfig(timeConfig *TimeConfig) *DateConfig {
	if timeConfig == nil {
		timeConfig = &TimeConfig{}
	}
	c := &DateConfig{TimeConfig: *timeConfig}
	c.Setup()
	return c
}

func (this *DateConfig) Setup() {
	if this.I18nKey == "" {
		this.I18nKey = I18NGROUP + ".metas.date.format"
	}
	if this.DefaultFormat == "" {
		this.DefaultFormat = "yyyy-MM-dd"
	}

	this.Type = TimeConfigDate
	this.TimeType = aorm.Date
}

// ConfigureQorMeta configure select one meta
func (this *DateConfig) ConfigureQorMeta(metaor resource.Metaor) {
	meta := metaor.(*Meta)
	if meta.Type == "" {
		meta.Type = "date"
	}
	this.Setup()
	this.TimeConfig.ConfigureQorMeta(metaor)
}

func (this *DateConfig) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Type == "" {
		filter.Type = "date"
	}
	this.Setup()
	this.TimeConfig.ConfigureQORAdminFilter(filter)
}
