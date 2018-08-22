package admin

import (
	"github.com/moisespsena-go/aorm"
	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
)

// Filter register filter for qor resource
func (res *Resource) Filter(filter *Filter) {
	filter.Resource = res

	if filter.Label == "" {
		filter.Label = utils.HumanizeString(filter.Name)
	}

	if filter.Config != nil {
		filter.Config.ConfigureQORAdminFilter(filter)
	}

	if filter.Handler == nil {
		// generate default handler
		filter.Handler = func(db *aorm.DB, filterArgument *FilterArgument) *aorm.DB {
			if metaValue := filterArgument.Value.Get("Value"); metaValue != nil {
				if keyword := utils.ToString(metaValue.Value); keyword != "" {
					field := filterField{FieldName: filter.Name}
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

					return filterResourceByFields(res, []filterField{field}, keyword, db, filterArgument.Context)
				}
			}
			return db
		}
	}

	if filter.Type != "" {
		res.filters[filter.Name] = filter
	} else {
		utils.ExitWithMsg("Invalid filter definition %v for resource %v", filter.Name, res.Name)
	}
}

func (res *Resource) GetFilters(context *core.Context) (filters []*Filter) {
	for _, filter := range res.filters {
		if filter.Available == nil || filter.Available(context) {
			filters = append(filters, filter)
		}
	}
	return
}

// Filter filter definiation
type Filter struct {
	Name       string
	Label      string
	Type       string
	Operations []string // eq, cont, gt, gteq, lt, lteq
	Resource   *Resource
	Handler    func(*aorm.DB, *FilterArgument) *aorm.DB
	Config     FilterConfigInterface
	Available func(context *core.Context) bool
}

// FilterConfigInterface filter config interface
type FilterConfigInterface interface {
	ConfigureQORAdminFilter(*Filter)
}

// FilterArgument filter argument that used in handler
type FilterArgument struct {
	Value    *resource.MetaValues
	Resource *Resource
	Context  *core.Context
}
