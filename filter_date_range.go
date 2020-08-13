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

func (this *FilterDateRange) GetValue(arg *FilterArgument) (pair interface{}, err error) {
	var times []time.Time
	if times, err = this.Parse(arg.Context, arg.Value.GetString("Start"), arg.Value.GetString("End")); err != nil {
		return
	}
	return TimeRange{times[0], times[1]}, nil
}

func (this *FilterDateRange) ConfigureQORAdminFilter(filter *Filter) {
	filter.Type = "date_range"
	if filter.Valuer == nil {
		filter.Valuer = this.GetValue
	}
}
