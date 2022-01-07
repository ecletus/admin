package admin

import (
	"strings"

	"github.com/moisespsena-go/xroute"

	"github.com/moisespsena-go/maps"

	"github.com/ecletus/core/utils"

	"github.com/ecletus/core/resource"

	"github.com/moisespsena-go/edis"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"
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
	Visible,
	ResetScopes,
	ResetFilters,
	ResetDefaultFilters,
	ResetSearchPrepareHandlers,
	ResetSearchCountHandlers,
	ResetSearchFindHandlers,
	ResetCountAggregationsHandlers bool
	Setup func(scheme *Scheme)
}

type Scheme struct {
	SchemeConfig    SchemeConfig
	EventDispatcher *SchemeDispatcher
	SchemeName      string
	SchemeParam     string
	Resource        *Resource

	SectionsAttribute

	sortableAttrs SortableAttrs

	SearchHandler SearchTermHandler

	SearchFindHandlers,
	SearchCountHandlers,
	PrepareSearchHandlers,
	CountAggregationsHandlers NamedSearcherHandlersRegistrator

	Scopes         ScopeRegistrator
	Filters        FilterRegistrator
	Categories     []string
	parentScheme   *Scheme
	Children       map[string]*Scheme
	Crumbs         core.BreadCrumberFunc
	DefaultFilters DBFilterRegistrator
	i18nKey        string
	NotMount       bool
	handler        *RouteHandler
	defaultMenu    *Menu

	itemMenus, menus []*Menu

	orders []interface{}

	hasFragments bool

	SchemeData maps.SyncedMap

	ApiHandlers map[string]func(ctx *Context)
}

func NewScheme(res *Resource, name string) *Scheme {
	s := &Scheme{
		Resource:                  res,
		SchemeName:                name,
		Scopes:                    &ScopeRegister{},
		Filters:                   &FilterRegister{},
		DefaultFilters:            &DBFilterRegister{},
		SearchCountHandlers:       &NamedSearcherHandlersRegistry{},
		SearchFindHandlers:        &NamedSearcherHandlersRegistry{},
		PrepareSearchHandlers:     &NamedSearcherHandlersRegistry{},
		CountAggregationsHandlers: &NamedSearcherHandlersRegistry{},
		SchemeParam:               utils.ToParamString(name),
		itemMenus:                 make([]*Menu, 0),
		ApiHandlers:               map[string]func(ctx *Context){},
		NotMount:                  res.Config.NotMount,
		SectionsAttribute: SectionsAttribute{
			Resource: res,
		},
	}

	if res.Scheme == nil {
		s.Sections = NewDefaultSchemeSectionsLayout(NewSchemeSectionsLayouts(name, NewSchemeSectionsLayoutsOptions{DefaultProvider: res.AllSectionsProvider}))
		s.AllSectionsFunc = res.AllSections
	} else {
		s.Sections = res.Scheme.Sections.MakeChild(name)
		s.AllSectionsFunc = res.Scheme.AllSectionsFunc
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
	if len(order) == 1 {
		if s, ok := order[0].([]string); ok {
			order = make([]interface{}, len(s))
			for i, s := range s {
				order[i] = s
			}
		}
	}
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

func (this *Scheme) CurrentSearchHandler() SearchTermHandler {
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

func (this *Scheme) ApplyDefaultFilters(ctx *Context) (_ *core.Context, err error) {
	db := ctx.DB()
	err = this.DefaultFilters.Each(map[string]*DBFilter{}, func(f *DBFilter) (err error) {
		db, err = f.Handler(ctx, db)
		return
	})
	if err == nil {
		ctx.SetRawDB(db)
		return ctx.Context, nil
	}
	return
}

func (this *Scheme) Breadcrumbs(ctx *core.Context) (crumbs []core.Breadcrumb, _ error) {
	if this == this.Resource.Scheme {
		return
	}
	if this.Crumbs != nil {
		return this.Crumbs(ctx)
	}
	crumbs = append(crumbs, core.NewBreadcrumb(this.Resource.GetIndexURI(ContextFromContext(ctx)), this.I18nKey()))
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

func (this *Scheme) GetCustomAttrs(name string) (Sections, bool) {
	panic("Scheme.GetCustomAttrs: deprecated")
}

func (this *Scheme) CustomAttrs(name string, values ...interface{}) Sections {
	return this.CustomAttrsOf(this.Sections.Default.Screen, name, values...)
}

func (this *Scheme) IndexAttrs(values ...interface{}) Sections {
	secs := this.IndexAttrsOf(this.Sections.Default.Screen, values...)
	this.SearchAttrs()
	return secs
}

func (this *Scheme) NewAttrs(values ...interface{}) Sections {
	return this.NewAttrsOf(this.Sections.Default.Screen, values...)
}

func (this *Scheme) EditAttrs(values ...interface{}) Sections {
	return this.EditAttrsOf(this.Sections.Default.Screen, values...)
}

func (this *Scheme) ShowAttrs(values ...interface{}) Sections {
	return this.ShowAttrsOf(this.Sections.Default.Screen, values...)
}

func (this *Scheme) NESAttrs(values ...interface{}) {
	this.NESAttrsOf(this.Sections.Default.Screen, values...)
}

func (this *Scheme) INESAttrs(values ...interface{}) {
	this.INESAttrsOf(this.Sections.Default.Screen, values...)
}

func (this *Scheme) IsSortableMeta(name string) bool {
	return this.sortableAttrs.Has(name)
}

// SortableAttrs set sortable attributes, sortable attributes could be click to order in qor table
func (this *Scheme) SortableAttrs(names ...string) []string {
	if len(names) != 0 || this.sortableAttrs.Names == nil {
		if len(names) == 0 {
			names = this.Sections.Default.Screen.Index.MetasNames()
		}
		this.sortableAttrs.Names = []string{}
		for _, name := range names {
			if field, ok := this.Resource.ModelStruct.FieldsByName[name]; ok && field.DBName != "" {
				this.sortableAttrs.Add(name)
			} else if m := this.Resource.GetMeta(name); m != nil && m.SortHandler != nil {
				this.sortableAttrs.Add(name)
			}
		}
	}
	return this.sortableAttrs.Names
}

func (this *Scheme) Unsort() {
	this.sortableAttrs.Names = nil
	this.sortableAttrs.Parent = nil
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (this *Scheme) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || this.SearchHandler == nil {
		if len(columns) == 0 {
			provider := this.Sections.Default.Screen.Index
			if !provider.IsSetI() && this != this.Resource.Scheme {
				return this.Resource.Scheme.SearchAttrs()
			} else {
				columns = provider.MustSections().MetasNames()
			}
		}

		if len(columns) > 0 {
			var filterFields []*filterField
			for _, column := range columns {
				parts := strings.SplitN(column, " ", 2)
				var op string
				if len(parts) == 2 {
					op, column = parts[0], parts[1]
				}
				f := NewFieldFilter(this.Resource, column)
				if f != nil {
					filterFields = append(filterFields, &filterField{Field: f, Operation: op})
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

func (this *Scheme) AddChild(name string, cfg ...*SchemeConfig) *Scheme {
	child := NewScheme(this.Resource, name)
	child.parentScheme = this
	var c = new(SchemeConfig)
	for _, c = range cfg {
	}

	if this.Children == nil {
		this.Children = map[string]*Scheme{}
	}

	if !c.ResetScopes {
		child.Scopes.InheritsFrom(this.Scopes)
	}

	if !c.ResetDefaultFilters {
		child.DefaultFilters.InheritsFrom(this.DefaultFilters)
	}

	if !c.ResetFilters {
		child.Filters.InheritsFrom(this.Filters)
	}

	if !c.ResetSearchCountHandlers {
		child.SearchCountHandlers.InheritsFrom(this.SearchCountHandlers)
	}

	if !c.ResetSearchFindHandlers {
		child.SearchFindHandlers.InheritsFrom(this.SearchFindHandlers)
	}

	if !c.ResetSearchPrepareHandlers {
		child.PrepareSearchHandlers.InheritsFrom(this.PrepareSearchHandlers)
	}

	if !c.ResetCountAggregationsHandlers {
		child.CountAggregationsHandlers.InheritsFrom(this.CountAggregationsHandlers)
	}

	child.sortableAttrs.Parent = &this.sortableAttrs

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

func (this *Scheme) GetItemMenusOf(ctx *Context, item interface{}) (menus []*Menu) {
	var all []*Menu
	if this.itemMenus == nil {
		if this.Resource.itemMenus != nil {
			all = this.Resource.itemMenus
		}
	}
	if all == nil {
		all = this.itemMenus
	}
	for _, m := range all {
		if m.ItemEnabled(ctx, item) {
			menus = append(menus, m)
		}
	}
	return
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

func (this *Scheme) ApiHandler(ext string, f func(ctx *Context)) {
	this.ApiHandlers[ext] = f
}
