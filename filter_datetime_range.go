package admin

import (
	"time"
)

type TimeRange struct {
	Start, End time.Time
}

type FilterDateTimeRange struct {
	DateTimeConfig
}

func (this *FilterDateTimeRange) GetValue(arg *FilterArgument) (pair interface{}, err error) {
	var times []time.Time
	if times, err = ParseTimeArgs(this.Layout(arg.Context), this.LocationFallbackValue(arg.Context),
		arg.Value.GetString("Start"), arg.Value.GetString("End")); err != nil {
		return
	}

	return TimeRange{times[0], times[1]}, nil
}

func (this *FilterDateTimeRange) ConfigureQORAdminFilter(filter *Filter) {
	filter.Type = "datetime_range"
	if filter.Valuer == nil {
		filter.Valuer = this.GetValue
	}
	this.DateTimeConfig.Configure()
}
