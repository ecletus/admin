package admin

import (
	"strings"

	"github.com/aghape/core"
	"github.com/aghape/roles"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-edis"
)

type Scheme struct {
	EventDispatcher edis.EventDispatcher
	SchemeName      string
	Resource        *Resource
	indexSections   []*Section
	newSections     []*Section
	editSections    []*Section
	isSetShowAttrs  bool
	showSections    []*Section
	customSections  *map[string]*[]*Section
	sortableAttrs   *[]string

	SearchHandler func(keyword string, context *core.Context) *aorm.DB

	scopes  []*Scope
	filters map[string]*Filter
}

// IndexAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (s *Scheme) IndexAttrs(values ...interface{}) []*Section {
	s.Resource.setSections(&s.indexSections, values...)
	s.SearchAttrs()
	return s.indexSections
}

// NewAttrs set attributes will be shown in the new page
//     // show given attributes in the new page
//     order.NewAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the new page
//     order.NewAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.NewAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (s *Scheme) NewAttrs(values ...interface{}) []*Section {
	s.Resource.setSections(&s.newSections, values...)
	return s.newSections
}

// EditAttrs set attributes will be shown in the edit page
//     // show given attributes in the new page
//     order.EditAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the edit page
//     order.EditAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.EditAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (s *Scheme) EditAttrs(values ...interface{}) []*Section {
	s.Resource.setSections(&s.editSections, values...)
	return s.editSections
}

// ShowAttrs set attributes will be shown in the show page
//     // show given attributes in the show page
//     order.ShowAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the show page
//     order.ShowAttrs("-State")
//  You could also use `Section` to structure form to make it tidy and clean
//     product.ShowAttrs(
//       &admin.Section{
//       	Title: "Basic Information",
//       	Rows: [][]string{
//       		{"Name"},
//       		{"Code", "Price"},
//       	}},
//       &admin.Section{
//       	Title: "Organization",
//       	Rows: [][]string{
//       		{"Category", "Collections", "MadeCountry"},
//       	}},
//       "Description",
//       "ColorVariations",
//     }
func (s *Scheme) ShowAttrs(values ...interface{}) []*Section {
	if len(values) > 0 {
		if values[len(values)-1] == false {
			values = values[:len(values)-1]
		} else {
			s.isSetShowAttrs = true
		}
	}
	s.Resource.setSections(&s.showSections, values...)
	return s.showSections
}

// SortableAttrs set sortable attributes, sortable attributes could be click to order in qor table
func (s *Scheme) SortableAttrs(columns ...string) []string {
	if len(columns) != 0 || s.sortableAttrs == nil {
		if len(columns) == 0 {
			columns = s.Resource.ConvertSectionToStrings(s.indexSections)
		}
		s.sortableAttrs = &[]string{}
		scope := core.FakeDB.NewScope(s.Resource.Value)
		for _, column := range columns {
			if field, ok := scope.FieldByName(column); ok && field.DBName != "" {
				attrs := append(*s.sortableAttrs, column)
				s.sortableAttrs = &attrs
			}
		}
	}
	return *s.sortableAttrs
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (s *Scheme) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || s.SearchHandler == nil {
		if len(columns) == 0 {
			if len(s.indexSections) == 0 && s != s.Resource.Scheme {
				return s.Resource.Scheme.SearchAttrs()
			} else {
				columns = s.Resource.ConvertSectionToStrings(s.indexSections)
			}
		}

		if len(columns) > 0 {
			s.SearchHandler = func(keyword string, context *core.Context) *aorm.DB {
				var filterFields []filterField
				for _, column := range columns {
					filterFields = append(filterFields, filterField{FieldName: column})
				}
				return filterResourceByFields(s.Resource, filterFields, keyword, context.DB, context)
			}
		}
	}

	return columns
}

func (s *Scheme) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return s.Resource.allAttrs()
	}

	var onlyExcludeAttrs = true
	for _, attr := range attrs {
		if !strings.HasPrefix(attr, "-") {
			onlyExcludeAttrs = false
			break
		}
	}

	if onlyExcludeAttrs {
		return append(s.Resource.allAttrs(), attrs...)
	}
	return attrs
}

func (s *Scheme) GetCustomAttrs(name string) ([]*Section, bool) {
	if s.customSections == nil {
		return nil, false
	}
	sections, ok := (*s.customSections)[name]
	if ok {
		return *sections, ok
	} else {
		return nil, false
	}
}

// CustomAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (s *Scheme) CustomAttrs(name string, values ...interface{}) []*Section {
	if s.customSections == nil {
		s.customSections = &map[string]*[]*Section{}
	}

	sections := &[]*Section{}
	s.Resource.setSections(sections, values...)
	(*s.customSections)[name] = sections

	return *sections
}

func (s *Scheme) IndexSections(context *Context) []*Section {
	if len(s.indexSections) == 0 && s.Resource.Scheme != s {
		return s.Resource.Scheme.IndexSections(context)
	}
	return s.Resource.allowedSections(nil, s.IndexAttrs(), context, roles.Read)
}

func (s *Scheme) EditSections(context *Context, record interface{}) []*Section {
	if len(s.editSections) == 0 && s.Resource.Scheme != s {
		return s.Resource.Scheme.EditSections(context, record)
	}
	return s.Resource.allowedSections(record, s.EditAttrs(), context, roles.Read)
}

func (s *Scheme) NewSections(context *Context) []*Section {
	if len(s.newSections) == 0 && s.Resource.Scheme != s {
		return s.Resource.Scheme.NewSections(context)
	}
	return s.Resource.allowedSections(nil, s.NewAttrs(), context, roles.Create)
}

func (s *Scheme) ShowSections(context *Context, record interface{}) []*Section {
	if len(s.showSections) == 0 && s.Resource.Scheme != s {
		return s.Resource.Scheme.ShowSections(context, record)
	}
	return s.Resource.allowedSections(record, s.ShowAttrs(), context, roles.Read)
}

func (s *Scheme) ContextSections(context *Context, recorde interface{}, action ...string) []*Section {
	var actio ContextType
	if len(action) > 0 && action[0] != "" {
		actio = ContextType(action[0])
	}
	switch actio {
	case NEW:
		return s.NewSections(context)
	case SHOW:
		return s.ShowSections(context, recorde)
	case EDIT:
		return s.EditSections(context, recorde)
	case INDEX:
		return s.IndexSections(context)
	}
	return nil
}
