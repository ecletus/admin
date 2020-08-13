package admin

import (
	"strings"

	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/maps"

	"github.com/ecletus/core/resource"

	"github.com/ecletus/core"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/edis"
)

const (
	E_SCHEME_ADDED      = "schemeAdded"
	AttrFragmentEnabled = "FragmentEnabled"
)

type SchemeDispatcher struct {
	edis.EventDispatcher
	Scheme *Scheme
}

type SchemeConfig struct {
	Visible bool
	ResetFilters,
	ResetDefaultFilters bool
	Setup func(scheme *Scheme)
}

type Scheme struct {
	SchemeConfig    SchemeConfig
	EventDispatcher *SchemeDispatcher
	SchemeName      string
	SchemeParam     string
	Resource        *Resource

	indexSections []*Section
	newSections   []*Section
	editSections  []*Section
	showSections  []*Section

	IndexSectionsFunc func(ctx *Context) []*Section
	NewSectionsFunc   func(ctx *Context) []*Section
	EditSectionsFunc  func(ctx *Context, record interface{}) []*Section
	ShowSectionsFunc  func(ctx *Context, record interface{}) []*Section

	isSetShowAttrs bool
	customSections *map[string]*[]*Section
	sortableAttrs  *[]string

	SearchHandler SearchHandler

	scopes             []*Scope
	Filters            FilterRegistrator
	Categories         []string
	parentScheme       *Scheme
	Children           map[string]*Scheme
	Crumbs             core.BreadCrumberFunc
	DefaultFilters     DBFilterRegistrator
	i18nKey            string
	NotMount           bool
	handler            *RouteHandler
	PrepareContextFunc func(ctx *core.Context)
	defaultMenu        *Menu

	itemMenus, menus []*Menu

	orders []interface{}

	hasFragments bool

	SchemeData maps.SyncedMap
}

func NewScheme(res *Resource, name string) *Scheme {
	s := &Scheme{
		Resource:       res,
		SchemeName:     name,
		Filters:        &FilterRegister{},
		DefaultFilters: &DBFilterRegister{},
		SchemeParam:    utils.ToParamString(name),
		itemMenus:      make([]*Menu, 0),
	}

	s.Filter(&Filter{
		Name:   "exclude",
		Hidden: true,
		Handler: func(db *aorm.DB, argument *FilterArgument) *aorm.DB {
			var keys []aorm.ID
			for _, v := range argument.Value.Values {
				if str := strings.TrimSpace(utils.ToString(v.Value)); str != "" {
					for _, v := range strings.Split(str, " ") {
						if key, err := s.Resource.ParseID(v); err != nil {
							argument.Context.AddError(err)
							return db
						} else {
							keys = append(keys, key)
						}
					}
				}
			}
			if keys == nil {
				return db
			}

			if sql, args, err := resource.IdToPrimaryQuery(argument.Context, s.Resource, true, keys...); err != nil {
				argument.Context.AddError(err)
			} else if sql != "" {
				return db.Where(sql, args...)
			}
			return db
		},
	})

	s.EventDispatcher = &SchemeDispatcher{Scheme: s}
	return s
}

func (this *Scheme) Order(order ...interface{}) *Scheme {
	this.orders = append(this.orders, order...)
	return this
}

func (this *Scheme) SetOrder(order ...interface{}) *Scheme {
	this.orders = order
	return this
}

func (this *Scheme) Orders() []interface{} {
	return this.orders
}

func (this *Scheme) CurrentOrders() []interface{} {
	if this.orders == nil {
		return this.Resource.Scheme.orders
	}
	return this.orders
}

func (this *Scheme) CurrentSearchHandler() SearchHandler {
	if this.SearchHandler == nil {
		return this.Resource.Scheme.SearchHandler
	}
	return this.SearchHandler
}

func (this *Scheme) DefaultMenu() *Menu {
	if this.defaultMenu == nil {
		if this == this.Resource.Scheme {
			this.defaultMenu = this.Resource.CreateMenu(!this.Resource.Config.Singleton)
		} else {
			this.defaultMenu = this.parentScheme.AddDefaultMenuChild(&Menu{
				Name: this.SchemeName,
				LabelFunc: func() string {
					return this.I18nKey()
				},
				URI: "/" + this.Resource.Param + this.Path(),
			})
		}
	}
	return this.defaultMenu
}

func (this *Scheme) AddDefaultMenuChild(child *Menu) *Menu {
	m := this.DefaultMenu()
	child.BaseResource = this.Resource
	m.subMenus = appendMenu(m.subMenus, nil, child)
	return child
}

func (this *Scheme) SetI18nKey(key string) *Scheme {
	this.i18nKey = key
	return this
}

func (this *Scheme) I18nKey() string {
	if this.i18nKey != "" {
		return this.i18nKey
	}
	return this.Resource.I18nPrefix + ".schemes." + this.SchemeName
}

func (this *Scheme) DefaultFilter(filter ...*DBFilter) {
	this.DefaultFilters.AddFilter(filter...)
}

func (this *Scheme) ApplyDefaultFilters(ctx *core.Context) (_ *core.Context, err error) {
	db := ctx.DB()
	err = this.DefaultFilters.Each(map[string]*DBFilter{}, func(f *DBFilter) (err error) {
		db, err = f.Handler(ctx, db)
		return
	})
	if err == nil {
		return ctx.SetRawDB(db), nil
	}
	return
}

func (this *Scheme) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb) {
	if this == this.Resource.Scheme {
		return
	}
	if this.Crumbs != nil {
		return this.Crumbs(ctx)
	}
	crumbs = append(crumbs, core.NewBreadcrumb(this.Resource.GetIndexURI(), this.I18nKey()))
	return
}

func (this *Scheme) Path() string {
	if this.parentScheme != nil && this.parentScheme.parentScheme != nil {
		return this.parentScheme.Path() + "/" + this.SchemeParam
	}
	return "/" + this.SchemeParam
}

func (this *Scheme) GetSchemeByName(param string) (child *Scheme) {
	parts := strings.Split(param, ".")

	if parts[0] == "" {
		parts = parts[1:]
	}

	child = this

	for _, p := range parts {
		if child.Children == nil {
			return nil
		}
		for _, c := range child.Children {
			if c.SchemeName == p {
				child = c
				break
			}
		}
	}
	return
}

func (this *Scheme) GetScheme(param string) (child *Scheme) {
	child, _ = this.GetSchemeOk(param)
	return
}

func (this *Scheme) GetSchemeOk(param string) (*Scheme, bool) {
	var ok bool
	if param == "" {
		return this, true
	}

	parts := strings.Split(param, ".")

	if parts[0] == "" {
		parts = parts[1:]
	}

	for _, p := range parts {
		if this.Children == nil {
			return nil, false
		}
		if this, ok = this.Children[p]; !ok {
			return nil, false
		}
	}
	return this, true
}

func (this *Scheme) OnDBActionE(cb func(e *resource.DBEvent) error, action ...resource.DBActionEvent) (err error) {
	return resource.OnDBActionE(this.EventDispatcher, cb, action...)
}

func (this *Scheme) OnDBAction(cb func(e *resource.DBEvent), action ...resource.DBActionEvent) (err error) {
	return resource.OnDBAction(this.EventDispatcher, cb, action...)
}

// IndexAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (this *Scheme) IndexAttrs(values ...interface{}) []*Section {
	this.Resource.setSections(&this.indexSections, values...)
	this.SearchAttrs()
	return this.indexSections
}

func (this *Scheme) excludeReadOnlyAttrs(at *[]*Section, values ...interface{}) []*Section {
	if len(values) == 0 {
		if len(*at) == 0 {
			// load defaults
			this.Resource.setSections(at)
		}
		values = append(values, *at)
		for _, f := range this.Resource.ModelStruct.ReadOnlyFields {
			values = append(values, "-"+f.Name)
		}
		// exclude readonly fields
		this.Resource.setSections(at, values...)
	} else {
		for _, f := range this.Resource.ModelStruct.ReadOnlyFields {
			values = append(values, "-"+f.Name)
		}

		this.Resource.setSections(at, values...)
	}
	return *at
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
func (this *Scheme) NewAttrs(values ...interface{}) []*Section {
	return this.excludeReadOnlyAttrs(&this.newSections, values...)
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
func (this *Scheme) EditAttrs(values ...interface{}) []*Section {
	return this.excludeReadOnlyAttrs(&this.editSections, values...)
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
func (this *Scheme) ShowAttrs(values ...interface{}) []*Section {
	if len(values) > 0 {
		if values[len(values)-1] == false {
			values = values[:len(values)-1]
		} else {
			this.isSetShowAttrs = true
		}
	} else {
		this.isSetShowAttrs = true
	}
	this.Resource.setSections(&this.showSections, values...)
	return this.showSections
}

func (this *Scheme) NESAttrs(values ...interface{}) {
	this.NewAttrs(values...)
	this.EditAttrs(values...)
	this.ShowAttrs(values...)
}

func (this *Scheme) INESAttrs(values ...interface{}) {
	this.IndexAttrs(values...)
	this.NewAttrs(values...)
	this.EditAttrs(values...)
	this.ShowAttrs(values...)
}

// SortableAttrs set sortable attributes, sortable attributes could be click to order in qor table
func (this *Scheme) SortableAttrs(columns ...string) []string {
	if len(columns) != 0 || this.sortableAttrs == nil {
		if len(columns) == 0 {
			columns = this.Resource.ConvertSectionToStrings(this.indexSections)
		}
		this.sortableAttrs = &[]string{}
		for _, column := range columns {
			if field, ok := this.Resource.ModelStruct.FieldsByName[column]; ok && field.DBName != "" {
				attrs := append(*this.sortableAttrs, column)
				this.sortableAttrs = &attrs
			}
		}
	}
	return *this.sortableAttrs
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (this *Scheme) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || this.SearchHandler == nil {
		if len(columns) == 0 {
			if len(this.indexSections) == 0 && this != this.Resource.Scheme {
				return this.Resource.Scheme.SearchAttrs()
			} else {
				columns = this.Resource.ConvertSectionToStrings(this.indexSections)
			}
		}

		if len(columns) > 0 {
			var filterFields []filterField
			for _, column := range columns {
				parts := strings.SplitN(column, " ", 2)
				var op string
				if len(parts) == 2 {
					op, column = parts[0], parts[1]
				}
				f := NewFieldFilter(this.Resource, column)
				if f != nil {
					filterFields = append(filterFields, filterField{Field: f, Operation: op})
				}
			}

			this.SearchHandler = func(searcher *Searcher, db *aorm.DB, keyword string) (*aorm.DB, error) {
				return filterResourceByFields(this.Resource, filterFields, keyword, db, searcher.Context.Context), nil
			}
		}
	}

	return columns
}

func (this *Scheme) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return this.Resource.allAttrs()
	}

	var onlyExcludeAttrs = true
	for _, attr := range attrs {
		if !strings.HasPrefix(attr, "-") {
			onlyExcludeAttrs = false
			break
		}
	}

	if onlyExcludeAttrs {
		return append(this.Resource.allAttrs(), attrs...)
	}
	return attrs
}

func (this *Scheme) GetCustomAttrs(name string) ([]*Section, bool) {
	if this.customSections == nil {
		return nil, false
	}
	sections, ok := (*this.customSections)[name]
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
func (this *Scheme) CustomAttrs(name string, values ...interface{}) []*Section {
	if this.customSections == nil {
		this.customSections = &map[string]*[]*Section{}
	}

	sections := &[]*Section{}
	this.Resource.setSections(sections, values...)
	(*this.customSections)[name] = sections

	return *sections
}

func (this *Scheme) IndexSections(context *Context) []*Section {
	if this.IndexSectionsFunc != nil {
		return this.IndexSectionsFunc(context)
	}

	if len(this.indexSections) == 0 {
		if this.parentScheme != nil {
			return this.parentScheme.IndexSections(context)
		}
		if this.Resource.Scheme != this {
			return this.Resource.Scheme.IndexSections(context)
		}
	}
	sections := this.Resource.allowedSections(nil, this.IndexAttrs(), context, roles.Read)
	return sections
}

func (this *Scheme) EditSections(context *Context, record interface{}) (sections []*Section) {
	if this.EditSectionsFunc != nil {
		return this.EditSectionsFunc(context, record)
	}

	if len(this.editSections) == 0 {
		if this.parentScheme != nil {
			return this.parentScheme.EditSections(context, record)
		}
		if this.Resource.Scheme != this {
			return this.Resource.Scheme.EditSections(context, record)
		}
	}
	if this == this.Resource.Scheme && this.Resource.Fragment != nil {
		sections = append(sections, &Section{Resource: this.Resource, Rows: [][]string{{AttrFragmentEnabled}}})
	}
	sections = append(sections, this.Resource.allowedSections(record, this.EditAttrs(), context, roles.Update)...)
	return sections
}

func (this *Scheme) NewSections(context *Context) []*Section {
	if this.NewSectionsFunc != nil {
		return this.NewSectionsFunc(context)
	}

	if len(this.newSections) == 0 {
		if this.parentScheme != nil {
			return this.parentScheme.NewSections(context)
		}
		if this.Resource.Scheme != this {
			return this.Resource.Scheme.NewSections(context)
		}
	}
	return this.Resource.allowedSections(nil, this.NewAttrs(), context, roles.Create)
}

func (this *Scheme) ShowSections(context *Context, record interface{}) []*Section {
	if this.ShowSectionsFunc != nil {
		return this.ShowSectionsFunc(context, record)
	}
	return this.ShowSectionsOriginal(context, record)
}

func (this *Scheme) ShowSectionsOriginal(context *Context, record interface{}) []*Section {
	if len(this.showSections) == 0 {
		if this.parentScheme != nil {
			return this.parentScheme.ShowSections(context, record)
		}
		if this.Resource.Scheme != this {
			return this.Resource.Scheme.ShowSections(context, record)
		}
	}
	return this.Resource.allowedSections(record, this.ShowAttrs(), context, roles.Read)
}

func (this *Scheme) ContextSections(context *Context, recorde interface{}, action ...string) []*Section {
	var b ContextType
	if len(action) > 0 && action[0] != "" {
		b = ParseContextType(action[0])
	} else {
		b = context.Type
	}

	if b.Has(NEW) {
		return this.NewSections(context)
	}
	if b.Has(SHOW) {
		return this.ShowSections(context, recorde)
	}
	if b.Has(EDIT) {
		return this.EditSections(context, recorde)
	}
	if b.Has(INDEX) {
		return this.IndexSections(context)
	}
	return nil
}

func (this *Scheme) Parents() (parents []*Scheme) {
	p := this.parentScheme
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

func (this *Scheme) PrepareContext(ctx *core.Context) {
	for _, p := range this.Parents() {
		if p.PrepareContextFunc != nil {
			p.PrepareContextFunc(ctx)
		}
	}

	if this.PrepareContextFunc != nil {
		this.PrepareContextFunc(ctx)
	}
}

func (this *Scheme) AddChild(name string, cfg ...*SchemeConfig) *Scheme {
	child := NewScheme(this.Resource, name)
	child.parentScheme = this
	var c = new(SchemeConfig)
	for _, c = range cfg {
	}

	if this.Children == nil {
		this.Children = map[string]*Scheme{}
	}

	if !c.ResetDefaultFilters {
		child.DefaultFilters.InheritsFrom(this.DefaultFilters)
	}

	if !c.ResetFilters {
		child.Filters.InheritsFrom(this.Filters)
	}

	if c.Setup != nil {
		c.Setup(child)
	}

	if !child.NotMount {
		if this.Resource.ControllerBuilder.Indexable() {
			child.handler = this.Resource.ControllerBuilder.ViewController.IndexHandler().Child()
		} else {
			child.handler = this.Resource.ControllerBuilder.ViewController.ReadHandler().Child()
		}
		child.Resource.Router.Get(child.Path(), child.handler)
		child.Resource.Router.Api(func(r xroute.Router) {
			r.Overrides(func(r xroute.Router) {
				r.Get(child.Path(), child.handler)
			})
		})

		if c.Visible {
			child.DefaultMenu().Permissioner = this.Resource
		}
	}

	this.Children[child.SchemeParam] = child
	this.Resource.triggerSchemeAdded(this)
	return child
}

// GetItemMenus get all sidebar itemMenus for admin
func (this *Scheme) GetItemMenus() (menus []*Menu) {
	if this.itemMenus == nil {
		if this.Resource.itemMenus != nil {
			return this.Resource.itemMenus
		}
		return
	}
	return this.itemMenus
}

// AddItemMenu add a menu to admin sidebar
func (this *Scheme) AddItemMenu(menu ...*Menu) *Menu {
	var m *Menu
	for _, m = range menu {
		m.prefix = this.Resource.Param
		menus := appendMenu(this.itemMenus, m.Ancestors, m)
		this.itemMenus = menus
	}
	return m
}

// GetItemMenu get sidebar menu with name
func (this *Scheme) GetItemMenu(name string) *Menu {
	var menus = this.itemMenus
	if menus == nil {
		menus = this.Resource.itemMenus
	}
	if menus == nil {
		return nil
	}

	return getMenu(menus, name)
}

// GetMenus get all sidebar itemMenus for admin
func (this *Scheme) GetMenus() (menus []*Menu) {
	if this.menus == nil {
		if this.Resource.menus != nil {
			return this.Resource.menus
		}
		return
	}
	return this.menus
}

// AddMenu add a menu to admin sidebar
func (this *Scheme) AddMenu(menu ...*Menu) *Menu {
	var m *Menu
	for _, m = range menu {
		m.prefix = this.Resource.Param
		menus := appendMenu(this.menus, m.Ancestors, m)
		this.menus = menus
	}
	return m
}

// GetMenu get sidebar menu with name
func (this *Scheme) GetMenu(name string) *Menu {
	var menus = this.menus
	if menus == nil {
		menus = this.Resource.menus
	}
	if menus == nil {
		return nil
	}

	return getMenu(menus, name)
}

type SchemeEvent struct {
	edis.EventInterface
	Scheme *Scheme
}
