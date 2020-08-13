package admin

import (
	"reflect"
	"sort"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"
)

type MetaFilterConfig struct {
	Resource *Resource
	Setup    func(meta *Meta, filter *Filter)
}

// MetaFilter register filter from resource meta
func (this *Scheme) MetaFilter(meta *Meta, setup func(meta *Meta, f *Filter)) *Filter {
	filter := ConvertMetaToFilter(meta, setup)
	this.Filter(filter)
	return filter
}

// MetaNameFilter register filter from resource meta by meta name
func (this *Scheme) MetaNameFilter(metaName string, cfg ...*MetaFilterConfig) *Filter {
	var (
		res = this.Resource
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
	this.Filter(filter)
	return filter
}

// MetaNamesFilter register filtersByName from resource meta by meta names
func (this *Scheme) MetaNamesFilter(metaNames []string, cfg ...*MetaFilterConfig) {
	for _, name := range metaNames {
		this.MetaNameFilter(name, cfg...)
	}
}

// Filter register filter for qor resource
func (this *Scheme) FilterByRelatedField(fieldName string, advanced ...bool) *Filter {
	meta := this.Resource.Meta(&Meta{Name: fieldName})
	filter := &Filter{
		Label: meta.Label,
	}
	for _, filter.advanced = range advanced{}
	if meta.Resource != nil {
		filter.Resource = meta.Resource
		filter.Type = "select_one"
		filter.Config = meta.Config.(*SelectOneConfig)
		filter.Name = fieldName
		field := meta.FieldStruct
		filter.Valuer = func(arg *FilterArgument) (value interface{}, err error) {
			var ID aorm.ID
			if ID, err = field.Relationship.ParseRelatedID(arg.Value.GetString("Value")); err != nil {
				return
			}
			ID = field.Relationship.ForeignID(ID)
			return ID, nil
		}
		filter.Handler = func(db *aorm.DB, arg *FilterArgument) *aorm.DB {
			return db.Where(arg.GoValue)
		}
		filter.LabelPairFunc = func(ctx *core.Context) (key, defaul string) {
			return meta.GetLabelPair()
		}
	}
	return this.Filter(filter)
}

// Filter register filter for qor resource
func (this *Scheme) Filter(filter *Filter) *Filter {
	filter.Resource = this.Resource
	filter.Scheme = this

	if filter.Label == "" && filter.LabelPairFunc == nil {
		filter.Label = utils.HumanizeString(filter.Name)
	}

	if filter.Type != "" {
		if setup, ok := FilterTypeSetup[filter.Type]; ok {
			setup(filter)
		}
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

			filter.Field = NewFieldFilter(this.Resource, fieldName)
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

						return filterResourceByFields(this.Resource, []filterField{field}, keyword, db, filterArgument.Context)
					}
				}
				return db
			}
		}
	}

	this.Filters.AddFilter(filter)
	return filter
}

func (this *Scheme) GetFilters(context *Context, advancedFilter ...bool) (filters []*Filter) {
	ok := func(f *Filter) bool {
		return true
	}
	if len(advancedFilter) > 0 {
		ok = func(f *Filter) bool {
			return f.advanced == advancedFilter[0]
		}
	}
	this.Filters.Each(map[string]*Filter{}, func(f *Filter) (err error) {
		if ok(f) && (f.Available == nil || f.Available(context)) {
			filters = append(filters, f)
		}
		return nil
	})
	return
}

func (this *Scheme) GetDefaultFilters(context *Context, advancedFilter ...bool) (filters []*Filter) {
	ok := func(f *Filter) bool {
		return true
	}
	if len(advancedFilter) > 0 {
		ok = func(f *Filter) bool {
			return f.advanced == advancedFilter[0]
		}
	}
	this.Filters.Each(map[string]*Filter{}, func(f *Filter) (err error) {
		if f.HandleEmpty && ok(f) && (f.Available == nil || f.Available(context)) {
			filters = append(filters, f)
		}
		return nil
	})
	return
}

func (this *Scheme) GetVisibleFilters(context *Context, advancedFilter ...bool) (filters []*Filter) {
	for _, filter := range this.GetFilters(context, advancedFilter...) {
		if filter.IsVisible(context) {
			filters = append(filters, filter)
		}
	}

	sort.Slice(filters, func(i, j int) bool {
		return filters[i].Name < filters[j].Name
	})

	return
}

func ConvertMetaToFilter(meta *Meta, cb ...func(meta *Meta, f *Filter)) (f *Filter) {
	f = &Filter{
		Name:  meta.Namer().Name,
		Label: meta.Label,
		LabelPairFunc: func(ctx *core.Context) (key, defaul string) {
			return meta.GetLabelPair()
		},
	}
	if meta.Config != nil {
		f.Config = meta.Config.(FilterConfigInterface)
	} else if meta.Config == nil {
		switch reflect.Indirect(reflect.New(meta.FieldStruct.Struct.Type)).Interface().(type) {
		case bool:
			f.Config = MetaConfigBooleanSelect().(FilterConfigInterface)
			f.Type = "select_one"
		}
	}
	if meta.FieldStruct != nil && meta.FieldStruct.BaseModel == meta.BaseResource.ModelStruct {
		f.FieldName = meta.Name
	}
	for _, cb := range cb {
		cb(meta, f)
	}
	return f
}

func ConvertMetaNameToFilter(res *Resource, metaName string, cb ...func(meta *Meta, f *Filter)) *Filter {
	return res.Filter(ConvertMetaToFilter(res.GetMetaOrSet(metaName), cb...))
}
