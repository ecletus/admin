package admin

import (
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

type MetaFilterConfig struct {
	Resource *Resource
	Setup    func(meta *Meta, filter *Filter)
}

// MetaFilter register filter from resource meta
func (s *Scheme) MetaFilter(meta *Meta, setup func(meta *Meta, f *Filter)) *Filter {
	filter := ConvertMetaToFilter(meta, setup)
	s.Filter(filter)
	return filter
}

// MetaNameFilter register filter from resource meta by meta name
func (s *Scheme) MetaNameFilter(metaName string, cfg ...*MetaFilterConfig) *Filter {
	var (
		res = s.Resource
		cb  func(meta *Meta, f *Filter)
	)
	if len(cfg) > 0 {
		if cfg[0].Resource != nil {
			res = cfg[0].Resource
		}
		if cfg[0].Setup != nil {
			cb = cfg[0].Setup
		}
	}
	filter := ConvertMetaNameToFilter(res, metaName, cb)
	s.Filter(filter)
	return filter
}

// MetaNamesFilter register filtersByName from resource meta by meta names
func (s *Scheme) MetaNamesFilter(metaNames []string, cfg ...*MetaFilterConfig) {
	for _, name := range metaNames {
		s.MetaNameFilter(name, cfg...)
	}
}

// Filter register filter for qor resource
func (s *Scheme) Filter(filter *Filter) *Filter {
	filter.Resource = s.Resource
	filter.Scheme = s

	if filter.Label == "" && filter.LabelFunc == nil {
		filter.Label = utils.HumanizeString(filter.Name)
	}

	if filter.Config != nil {
		filter.Config.ConfigureQORAdminFilter(filter)
	}

	if filter.Handler == nil {
		if filter.Field == nil {
			fieldName := filter.FieldName
			if fieldName == "" {
				fieldName = filter.Name
			}

			filter.Field = NewFieldFilter(s.Resource, fieldName)
		}

		if filter.Field == nil {
			filter.Handler = func(db *aorm.DB, filterArgument *FilterArgument) *aorm.DB {
				return db
			}
		} else {
			// generate default handler
			filter.Handler = func(db *aorm.DB, filterArgument *FilterArgument) *aorm.DB {
				if metaValue := filterArgument.Value.Get("Value"); metaValue != nil {
					if keyword := utils.ToString(metaValue.Value); keyword != "" {
						field := filterField{
							Field:     filter.Field,
							Operation: filterArgument.Filter.DefaultOperation,
						}
						if operationMeta := filterArgument.Value.Get("Operation"); operationMeta != nil {
							if operation := utils.ToString(operationMeta.Value); operation != "" {
								field.Operation = operation
							}
						}
						if field.Operation == "" {
							if len(filter.Operations) > 0 {
								field.Operation = filter.Operations[0]
							} else {
								field.Operation = "contains"
							}
						}

						return filterResourceByFields(s.Resource, []filterField{field}, keyword, db, filterArgument.Context)
					}
				}
				return db
			}
		}
	}

	s.filters = append(s.filters, filter)
	s.filtersByName[filter.Name] = filter
	return filter
}

func (s *Scheme) GetFilters(context *Context, advancedFilter ...bool) (filters []*Filter) {
	ok := func(f *Filter) bool {
		return true
	}
	if len(advancedFilter) > 0 {
		ok = func(f *Filter) bool {
			return f.advanced == advancedFilter[0]
		}
	}
	for _, filter := range s.filters {
		if ok(filter) && (filter.Available == nil || filter.Available(context)) {
			filters = append(filters, filter)
		}
	}
	return
}

func (s *Scheme) GetVisibleFilters(context *Context, advancedFilter ...bool) (filters []*Filter) {
	for _, filter := range s.GetFilters(context, advancedFilter...) {
		if filter.IsVisible(context) {
			filters = append(filters, filter)
		}
	}
	return
}

func ConvertMetaToFilter(meta *Meta, cb func(meta *Meta, f *Filter)) (f *Filter) {
	f = &Filter{
		Name:      meta.Namer().Name,
		Label:     meta.Label,
		LabelFunc: meta.GetLabelPair,
		Config:    meta.Config.(FilterConfigInterface),
	}

	if cb != nil {
		cb(meta, f)
	}
	return f
}

func ConvertMetaNameToFilter(res *Resource, metaName string, cb func(meta *Meta, f *Filter)) *Filter {
	return ConvertMetaToFilter(res.GetMetaOrSet(metaName), cb)
}
