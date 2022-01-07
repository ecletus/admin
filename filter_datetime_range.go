package admin

type TimeRange struct {
	Start, End TimeValue
}

func (t TimeRange) UTC() TimeRange {
	return TimeRange{t.Start.UTC(), t.End.UTC()}
}

func (t TimeRange) Date() TimeRange {
	return TimeRange{t.Start.ToDate(), t.End.ToDate()}
}

func (t TimeRange) TimeStamp() TimeRange {
	return TimeRange{t.Start.ToTimeStamp(), t.End.ToTimeStamp()}
}

func (t TimeRange) TimeStampTz() TimeRange {
	return TimeRange{t.Start.ToTimeStampTz(), t.End.ToTimeStampTz()}
}

func (t TimeRange) Time() TimeRange {
	return TimeRange{t.Start.ToTime(), t.End.ToTime()}
}

func (t TimeRange) TimeTz() TimeRange {
	return TimeRange{t.Start.ToTimeTz(), t.End.ToTimeTz()}
}

func (t TimeRange) IsZero() bool {
	return t.Start.IsZero() && t.End.IsZero()
}

type FilterDateTimeRange struct {
	DateTimeConfig
}

func (this *FilterDateTimeRange) ConfigureQORAdminFilter(filter *Filter) {
	if filter.Type == "" {
		filter.Type = "datetime_range"
	}
	this.DateTimeConfig.ConfigureQORAdminFilter(filter)
}
