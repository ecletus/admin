package admin

var FilterTypeSetup = map[string]func(filter *Filter){
	"datetime_range": func(filter *Filter) {
		if filter.Config == nil {
			filter.Config = &FilterDateTimeRange{}
		}
		filter.Config.ConfigureQORAdminFilter(filter)
	},
	"date_range": func(filter *Filter) {
		if filter.Config == nil {
			filter.Config = &FilterDateRange{}
		}
		filter.Config.ConfigureQORAdminFilter(filter)
	},
}
