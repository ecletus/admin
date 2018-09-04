package admin

import (
	"strings"

	"github.com/aghape/core/utils"

	"github.com/aghape/core/resource"

	"github.com/aghape/core"
	"github.com/aghape/roles"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-edis"
)

const E_SCHEME_ADDED = "schemeAdded"

type SchemeDispatcher struct {
	edis.EventDispatcher
	Scheme *Scheme
}

type SchemeConfig struct {
	Visible bool
	Setup func(scheme *Scheme)
}

type Scheme struct {
	SchemeConfig SchemeConfig
	EventDispatcher *SchemeDispatcher
	SchemeName      string
	SchemeParam     string
	Resource        *Resource
	indexSections   []*Section
	newSections     []*Section
	editSections    []*Section
	isSetShowAttrs  bool
	showSections    []*Section
	customSections  *map[string]*[]*Section
	sortableAttrs   *[]string

	SearchHandler func(keyword string, context *core.Context) *aorm.DB

	scopes             []*Scope
	filters            map[string]*Filter
	Categories         []string
	parentScheme       *Scheme
	Children           map[string]*Scheme
	Crumbs             core.BreadcrumberFunc
	DefaultFilters     []func(context *core.Context, db *aorm.DB) *aorm.DB
	i18nKey            string
	NotMount           bool
	handler            *RouteHandler
	PrepareContextFunc func(ctx *core.Context)
	defaultMenu        *Menu
}

func (s *Scheme) DefaultMenu() *Menu {
	if s.defaultMenu == nil {
		if s == s.Resource.Scheme {
			s.defaultMenu = s.Resource.CreateMenu(!s.Resource.Config.Singleton)
		} else {
			s.defaultMenu = s.parentScheme.AddDefaultMenuChild(&Menu{
				Name:s.SchemeName,
				LabelFunc: func() string {
					return s.I18nKey()
				},
				RelativePath:"/" + s.Resource.Param + s.Path(),
			})
		}
	}
	return s.defaultMenu
}

func (s *Scheme) AddDefaultMenuChild(child *Menu) *Menu {
	m := s.DefaultMenu()
	if len(m.subMenus) == 0 {
		c := *m
		m.Resource = nil
		m.Permissioner = nil
		c.Name += "All"
		m.subMenus = appendMenu(m.subMenus, m.Ancestors, &c)
	}
	m.subMenus = appendMenu(m.subMenus, m.Ancestors, child)
	return child
}

func (s *Scheme) SetI18nKey(key string) *Scheme {
	s.i18nKey = key
	return s
}

func (s *Scheme) I18nKey() string {
	if s.i18nKey != "" {
		return s.i18nKey
	}
	return s.Resource.I18nPrefix + ".schemes." + s.SchemeName
}

func (s *Scheme) DefaultFilter(fns ...func(context *core.Context, db *aorm.DB) *aorm.DB) {
	s.DefaultFilters = append(s.DefaultFilters, fns...)
}

func (s *Scheme) ApplyDefaultFilters(context *core.Context) *core.Context {
	if s.DefaultFilters == nil {
		return context
	}
	context = context.Clone()
	db := context.DB
	for _, df := range s.DefaultFilters {
		db = df(context, db)
	}
	context.SetDB(db)
	return context
}

func (s *Scheme) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb) {
	if s == s.Resource.Scheme {
		return
	}
	if s.Crumbs != nil {
		return s.Crumbs(ctx)
	}
	crumbs = append(crumbs, core.NewBreadcrumb(s.Resource.GetIndexURI(), s.I18nKey()))
	return
}

func (s *Scheme) Path() string {
	if s.parentScheme != nil && s.parentScheme.parentScheme != nil {
		return s.parentScheme.Path() + "/" + s.SchemeParam
	}
	return "/" + s.SchemeParam
}

func (s *Scheme) GetScheme(param string) (child *Scheme, ok bool) {
	if param == "" {
		return s, true
	}

	parts := strings.Split(param, ".")

	if parts[0] == "" {
		parts = parts[1:]
	}

	for _, p := range parts {
		if s.Children == nil {
			return nil, false
		}
		if s, ok = s.Children[p]; !ok {
			return nil, false
		}
	}
	return s, true
}

func NewScheme(res *Resource, name string) *Scheme {
	s := &Scheme{
		Resource:    res,
		SchemeName:  name,
		filters:     make(map[string]*Filter),
		SchemeParam: utils.ToParamString(name),
	}
	s.EventDispatcher = &SchemeDispatcher{Scheme: s}
	return s
}

func (s *Scheme) OnDBActionE(cb func(e *resource.DBEvent) error, action ...resource.DBActionEvent) (err error) {
	return resource.OnDBActionE(s.EventDispatcher, cb, action...)
}

func (s *Scheme) OnDBAction(cb func(e *resource.DBEvent), action ...resource.DBActionEvent) (err error) {
	return resource.OnDBAction(s.EventDispatcher, cb, action...)
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
	} else {
		actio = context.Type
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

func (s *Scheme) Parents() (parents []*Scheme) {
	p := s.parentScheme
	for p != nil {
		parents = append(parents, p)
		p = p.parentScheme
	}
	l := len(parents)
	for i := l/2 - 1; i >= 0; i-- {
		opp := l - 1 - i
		parents[i], parents[opp] = parents[opp], parents[i]
	}
	return
}

func (s *Scheme) PrepareContext(ctx *core.Context) {
	for _, p := range s.Parents() {
		if p.PrepareContextFunc != nil {
			p.PrepareContextFunc(ctx)
		}
	}

	if s.PrepareContextFunc != nil {
		s.PrepareContextFunc(ctx)
	}
}

func (s *Scheme) AddChild(name string, cfg ...*SchemeConfig) *Scheme {
	child := NewScheme(s.Resource, name)
	child.parentScheme = s
	var c *SchemeConfig
	if len(cfg) > 0 {
		c = cfg[0]
	}

	if s.Children == nil {
		s.Children = map[string]*Scheme{}
	}

	if c.Setup != nil {
		c.Setup(child)
	}

	if !child.NotMount {
		child.handler = s.Resource.IndexHandler().Child()
		child.Resource.Router.Get(child.Path(), child.handler)

		if c.Visible {
			child.DefaultMenu()
		}
	}

	s.Children[child.SchemeParam] = child
	s.Resource.triggerSchemeAdded(s)
	return child
}

func (s *Scheme) IndexIntersept(interseptor func(chain *Chain)) {
	s.Resource.IndexHandler().Intercept(interseptor)
}

type SchemeEvent struct {
	edis.EventInterface
	Scheme *Scheme
}
