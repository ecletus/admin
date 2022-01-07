package admin

import (
	"time"
)

type DateRange struct {
	Start, End time.Time
}

type FilterDateRange struct {
	DateConfig
}

func (this *FilterDateRange) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Type == "" {
		filter.Type = "date_range"
	}
	this.DateConfig.ConfigureQORAdminFilter(filter)
}
