package admin

var FilterTypeSetup = map[string]func(filter *Filter){
	"datetime_range": func(filter *Filter) {
		if filter.Config == nil {
			filter.Config = &FilterDateTimeRange{}
		}
	},
	"date_range": func(filter *Filter) {
		if filter.Config == nil {
			filter.Config = &FilterDateRange{}
		}
	},
}
