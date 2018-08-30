package admin

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/moisespsena/go-edis"

	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/aghape/helpers"
	"github.com/aghape/roles"
	"github.com/jinzhu/inflection"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena/go-route"

	//"github.com/aghape/responder"
	"strconv"

	"github.com/aghape/db/inheritance"
	"github.com/aghape/fragment"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/template/html/template"
)

const (
	DEFAULT_LAYOUT = resource.DEFAULT_LAYOUT

	// paths
	P_NEW_FORM              = "/new"
	P_NEW                   = "/"
	P_OBJ_READ              = "/"
	P_OBJ_READ_FORM         = "/"
	P_OBJ_UPDATE            = "/"
	P_OBJ_UPDATE_FORM       = "/edit"
	P_OBJ_DELETE            = "/"
	P_SINGLETON_READ        = P_OBJ_READ
	P_SINGLETON_READ_FORM   = P_OBJ_READ_FORM
	P_SINGLETON_UPDATE      = P_OBJ_UPDATE
	P_SINGLETON_UPDATE_FORM = P_OBJ_UPDATE_FORM
	P_INDEX                 = "/"

	// actions
	A_CREATE = "create"
	A_UPDATE = "update"
	A_READ   = "read"
	A_DELETE = "delete"
	A_INDEX  = "index"
)

type SubResourceConfig struct {
	Value         interface{}
	FieldName     string
	LabelPlural   string
	LabelSingular string
	IconSingular  string
	IconPlural    string
	Invisible     bool
	MenuEnabled   func(record interface{}, context *Context)
}

type SubResource struct {
	Resource *Resource
	Config   *SubResourceConfig
}

func (res *SubResource) CreateMenu(plural bool, parentParams ...string) *Menu {
	menu := res.Resource.CreateMenu(plural)
	if plural {
		if res.Config.LabelPlural != "" {
			menu.Label = res.Config.LabelPlural
		}
		if res.Config.IconPlural != "" {
			menu.Icon = res.Config.IconPlural
		}
	} else {
		if res.Config.LabelSingular != "" {
			menu.Label = res.Config.LabelSingular
		}
		if res.Config.IconSingular != "" {
			menu.Icon = res.Config.IconSingular
		}
	}

	menu.RelativePath = res.Resource.GetIndexURI(parentParams...)
	return menu
}

func (res *SubResource) CreateDefaultMenu(parentParams ...string) *Menu {
	return res.CreateMenu(!res.Resource.Config.Singleton, parentParams...)
}

// Resource is the most important thing for qor admin, every model is defined as a resource, qor admin will genetate management interface based on its definition
type Resource struct {
	*resource.Resource
	*Scheme
	ParentResource   *Resource
	Config           *Config
	Metas            []*Meta
	MetasByName      map[string]*Meta
	MetasByFieldName map[string]*Meta
	SingleEditMetas  map[string]*Meta
	Actions          []*Action

	admin           *Admin
	mounted         bool
	cachedMetas     *map[string][]*Meta
	AdminController *Controller

	Router       *route.Mux
	ObjectRouter *route.Mux
	Parents      []*Resource
	Param        string
	ParamName    string
	paramIDName  string

	Resources          map[string]*Resource
	ResourcesByParam   map[string]*Resource
	menus              []*Menu
	MetaAliases        map[string]*resource.MetaName
	defaultDisplayName string
	Children           *Inheritances
	Inherits           map[string]*Child
	Fragments          *Fragments
	Fragment           *Fragment
	DefaultMenu        *Menu
	registered         bool
	afterRegister      []func()
}

func (res *Resource) IndexHandler() *RouteHandler {
	return res.Router.FindHandler("GET", P_INDEX).(*RouteHandler)
}

func (res *Resource) OnDBActionE(cb func(e *resource.DBEvent) error, action ...resource.DBActionEvent) (err error) {
	return resource.OnDBActionE(res, cb, action...)
}

func (res *Resource) OnDBAction(cb func(e *resource.DBEvent), action ...resource.DBActionEvent) (err error) {
	return resource.OnDBAction(res, cb, action...)
}

// GetMenus get all sidebar menus for admin
func (res *Resource) GetMenus() []*Menu {
	return res.menus
}

// AddMenu add a menu to admin sidebar
func (res *Resource) AddMenu(menu *Menu) *Menu {
	menu.router = res.Router
	res.menus = appendMenu(res.menus, menu.Ancestors, menu)
	return menu
}

// GetMenu get sidebar menu with name
func (res *Resource) GetMenu(name string) *Menu {
	return getMenu(res.menus, name)
}

// GetDBKeys Returns the DB Keys values from `value`.
func (res *Resource) GetDBKeys(value interface{}) (keys []string) {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	for _, field := range res.FakeScope.PrimaryFields() {
		keys = append(keys, fmt.Sprint(reflectValue.FieldByIndex(field.Struct.Index).Interface()))
	}
	return
}

func (res *Resource) GetDBKey(value interface{}) string {
	return strings.Join(res.GetDBKeys(value), ",")
}

// GetKeys Returns the Resource Keys values from `value`.
func (res *Resource) GetKeys(value interface{}) (keys []string) {
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	for _, field := range res.PrimaryFields {
		rf := reflectValue.FieldByName(field.Struct.Name).Interface()
		keys = append(keys, fmt.Sprint(rf))
	}
	return keys
}

func (res *Resource) GetKey(value interface{}) (key string) {
	switch vt := value.(type) {
	case interface{ GetID() string }:
		return vt.GetID()
	case interface{ GetID() int64 }:
		return fmt.Sprint(vt.GetID())
	default:
		return strings.Join(res.GetKeys(value), ",")
	}
}

func (res *Resource) LabelKey() string {
	return res.I18nPrefix + ".label"
}

func (res *Resource) GetLabelKey(plural bool) string {
	r := res.LabelKey() + "~"
	if plural {
		r += "p"
	} else {
		r += "s"
	}
	return r
}

func (res *Resource) PluralLabelKey() string {
	return res.GetLabelKey(true)
}

func (res *Resource) SingularLabelKey() string {
	return res.GetLabelKey(false)
}

func (res *Resource) GetDefaultLabel(plural bool) string {
	if plural {
		return inflection.Plural(res.Name)
	} else {
		return res.Name
	}
}

func (res *Resource) GetLabel(context *Context, plural bool) string {
	return string(context.t(res.GetLabelKey(plural), res.GetDefaultLabel(plural)))
}

func (res *Resource) GetActionLabelKey(action *Action) string {
	return fmt.Sprintf("resources.%v.actions.%v", res.ToParam(), action.Label)
}

func (res *Resource) GetActionLabel(context *Context, action *Action) template.HTML {
	return context.t(res.GetActionLabelKey(action), action.Label)
}

// Meta register meta for admin resource
func (res *Resource) SetMeta(meta *Meta, notUpdate ...bool) *Meta {
	return res.Meta(meta, true)
}

// Meta register meta for admin resource
func (res *Resource) Meta(meta *Meta, notUpdate ...bool) *Meta {
	if oldMeta := res.GetMeta(meta.Name, notUpdate...); oldMeta != nil {
		if meta.Type != "" {
			oldMeta.Type = meta.Type
			oldMeta.Config = nil
		}

		if meta.TypeHandler != nil {
			oldMeta.TypeHandler = meta.TypeHandler
		}

		if meta.Enabled != nil {
			oldMeta.Enabled = meta.Enabled
		}

		if meta.SkipDefaultLabel {
			oldMeta.SkipDefaultLabel = true
		}

		if meta.DefaultLabel != "" {
			oldMeta.DefaultLabel = meta.DefaultLabel
		}

		if meta.Label != "" {
			oldMeta.Label = meta.Label
		}

		if meta.FieldName != "" {
			oldMeta.FieldName = meta.FieldName
		}

		if meta.Setter != nil {
			oldMeta.Setter = meta.Setter
		}

		if meta.Valuer != nil {
			oldMeta.Valuer = meta.Valuer
		}

		if meta.FormattedValuer != nil {
			oldMeta.FormattedValuer = meta.FormattedValuer
		}

		if meta.Resource != nil {
			oldMeta.Resource = meta.Resource
		}

		if meta.Permission != nil {
			oldMeta.Permission = meta.Permission
		}

		if meta.Config != nil {
			oldMeta.Config = meta.Config
		}

		if meta.Collection != nil {
			oldMeta.Collection = meta.Collection
		}

		if meta.EditName != "" {
			if meta.EditName == "-#" {
				meta.EditName = strings.TrimSuffix(meta.Name, "ID")
			}
			oldMeta.EditName = meta.EditName
		}

		if len(meta.Dependency) > 0 {
			oldMeta.Dependency = meta.Dependency
		}

		if meta.Fragment != nil {
			oldMeta.Fragment = meta.Fragment
		}

		meta = oldMeta
	} else {
		res.MetasByName[meta.Name] = meta
		res.Metas = append(res.Metas, meta)
		meta.baseResource = res
	}

	meta.updateMeta()
	return meta
}

func (res *Resource) InitRoutes() *route.Mux {
	if !res.Config.Singleton {
		for param, subRes := range res.ResourcesByParam {
			r := subRes.InitRoutes()
			pattern := "/" + param
			res.ObjectRouter.Mount(pattern, r)
		}
		res.Router.Mount("/"+res.ParamIDPattern(), res.ObjectRouter)
	}
	return res.Router
}

// GetAdmin get admin from resource
func (res *Resource) GetAdmin() *Admin {
	return res.admin
}

// GetURL
func (res *Resource) GetIndexURI(parentkeys ...string) string {
	var p []string
	r := res.ParentResource
	for l, i := len(parentkeys), 0; i < l; i++ {
		p = append(p, r.ToParam(), parentkeys[i])
		r = res.ParentResource
	}
	return "/" + strings.Join(append(p, res.ToParam()), "/")
}

// GetURL
func (res *Resource) GetURI(key string, parentkeys ...string) string {
	return res.GetIndexURI(parentkeys...) + "/" + key
}

// GetURL
func (res *Resource) GetContextIndexURI(context *core.Context, parentkeys ...string) string {
	var p []string
	if len(parentkeys) == 0 {
		if res.ParentResource != nil {
			return res.ParentResource.GetContextURI(context, "") + "/" + res.ToParam()
		}
	} else {
		r := res.ParentResource
		for l, i := len(parentkeys), 0; i < l; i++ {
			p = append(p, r.ToParam(), parentkeys[i])
			r = res.ParentResource
		}
	}
	return context.GenURL(append(p, res.ToParam())...)
}

// GetURL
func (res *Resource) GetContextURI(context *core.Context, key string, parentkeys ...string) string {
	if key == "" {
		key = context.URLParam(res.ParamIDName())
	}
	return res.GetContextIndexURI(context, parentkeys...) + "/" + key
}

func (res *Resource) GetRecordURI(record interface{}, parentKeys ...string) string {
	if ref, ok := record.(inheritance.ParentModelInterface); ok {
		if child := ref.GetQorChild(); child != nil {
			res = res.Children.Items[child.Index].Resource
			// TODO Fix support for subresources
			return res.GetURI(child.ID)
		}
	}
	return res.GetURI(res.GetKey(record), parentKeys...)
}

// GetPrimaryValue get priamry value from request
func (res *Resource) GetPrimaryValue(params utils.ReadonlyMapString) string {
	if params != nil {
		return params.Get(res.ParamIDName())
	}
	return ""
}

// ParamIDName return param name for primary key like :product_id
func (res *Resource) ParamIDName() string {
	return res.paramIDName
}

// ParamIDName return param name for primary key like :product_id
func (res *Resource) ParamIDPattern() string {
	return "{" + res.paramIDName + "}"
}

// ToParam used as urls to register routes for resource
func (res *Resource) ToParam() string {
	return res.Param
}

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (res *Resource) UseDisplay(display interface{}) {
	var displayInterface DisplayInterface
	if ti, ok := display.(DisplayInterface); ok {
		displayInterface = ti
	} else if str, ok := display.(string); ok {
		if res.GetDisplay(str) != nil {
			return
		}

		displayInterface = &Display{Name: str}
	}

	if displayInterface != nil {
		if res.Config.Displays == nil {
			res.Config.Displays = make(map[string]DisplayInterface)
		}
		res.Config.Displays[displayInterface.GetName()] = displayInterface
		displayInterface.ConfigAdminTheme(res)
	}
}

func (res *Resource) GetDefaultDisplayName() string {
	if res.defaultDisplayName == "" {
		return "default"
	}
	return res.defaultDisplayName
}

func (res *Resource) SetDefaultDisplay(displayName string) {
	display := res.GetDisplay(displayName)
	if display == nil {
		panic(fmt.Errorf("Display %q does not exists.", displayName))
	}
	res.defaultDisplayName = displayName
}

func (res *Resource) GetDefaultDisplay() DisplayInterface {
	display := res.GetDisplay(res.GetDefaultDisplayName())
	if display == nil {
		return DefaultDisplay
	}
	return display
}

// GetDisplay get registered theme with name
func (res *Resource) GetDisplay(name string) DisplayInterface {
	if res.Config.Displays != nil {
		if d, ok := res.Config.Displays[name]; ok {
			return d
		}
	}
	return nil
}

// UseTheme use them for resource, will auto load the theme's javascripts, stylesheets for this resource
func (res *Resource) UseTheme(theme interface{}) []ThemeInterface {
	var themeInterface ThemeInterface
	if ti, ok := theme.(ThemeInterface); ok {
		themeInterface = ti
	} else if str, ok := theme.(string); ok {
		for _, theme := range res.Config.Themes {
			if theme.GetName() == str {
				return res.Config.Themes
			}
		}

		themeInterface = Theme{Name: str}
	}

	if themeInterface != nil {
		res.Config.Themes = append(res.Config.Themes, themeInterface)
		themeInterface.ConfigAdminTheme(res)
	}
	return res.Config.Themes
}

// GetTheme get registered theme with name
func (res *Resource) GetTheme(name string) ThemeInterface {
	for _, theme := range res.Config.Themes {
		if theme.GetName() == name {
			return theme
		}
	}
	return nil
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
func (res *Resource) NewResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = res
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}
	return res.admin.NewResource(value, config[0])
}

// AddSubResource register sub-resource
func (res *Resource) AddResource(cfg *SubConfig, value interface{}, config ...*Config) *Resource {
	cfg.Parent = res
	if len(config) == 0 {
		config = []*Config{{Sub: cfg}}
	} else {
		config[0].Sub = cfg
	}
	return res.AddResourceConfig(value, config[0])
}

// AddSubResource register sub-resource
func (res *Resource) AddResourceConfig(value interface{}, cfg *Config) *Resource {
	if cfg.Sub == nil {
		cfg.Sub = &SubConfig{}
	}
	cfg.Sub.Parent = res
	return res.admin.AddResource(value, cfg)
}

func (res *Resource) DBName(quote bool) (name string, alias string, pkfields []string) {
	if quote {
		name = res.FakeScope.QuotedTableName()
	} else {
		name = res.FakeScope.TableName()
	}

	alias = "parent_" + strconv.Itoa(res.PathLevel)
	for _, f := range res.PrimaryFields {
		pkfields = append(pkfields, f.DBName)
	}

	return
}

func (res *Resource) FilterByParent(db *aorm.DB, parentKey string) *aorm.DB {
	r := res.ParentResource

	res_parent_alias := res.FakeScope.QuotedTableName()
	res_pkfield := res.ParentFieldDBName

	var (
		parent_name, parent_alias string
		pkfield                   []string
	)

	parent_name, parent_alias, pkfield = r.DBName(true)
	parent_alias += "_"
	var fields, wheres []string
	ids := strings.Split(parentKey, ",")

	fields = append(fields, fmt.Sprintf("%[1]v.%[2]v = %[3]v.%[4]v", parent_alias,
		pkfield[0], res_parent_alias, res_pkfield))
	wheres = append(wheres, parent_alias+"."+pkfield[0]+" = ?")

	join := fmt.Sprintf("JOIN %v as %v ON %v", parent_name, parent_alias,
		strings.Join(fields, " AND "))

	idsinterface := make([]interface{}, len(ids), len(ids))
	for i, id := range ids {
		idsinterface[i] = id
	}

	db = db.Joins(join).Where(strings.Join(wheres, " AND "), idsinterface...)

	r = r.ParentResource
	return db
}

func (res *Resource) GetPathLevel() int {
	return res.PathLevel
}

// Decode decode context into a value
func (res *Resource) Decode(context *core.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &aorm.Scope{Value: res.Value}

Fields:
	for _, field := range scope.GetModelStruct().StructFields {
		for _, meta := range res.Metas {
			if field.Name == meta.FieldName {
				attrs = append(attrs, meta.Name)
				continue Fields
			}
		}

		if field.IsForeignKey {
			continue
		}

		for _, value := range []string{"CreatedAt", "UpdatedAt", "DeletedAt"} {
			if value == field.Name {
				continue Fields
			}
		}

		if (field.IsNormal || field.Relationship != nil) && !field.IsIgnored {
			attrs = append(attrs, field.Name)
			continue
		}

		fieldType := field.Struct.Type
		for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			attrs = append(attrs, field.Name)
		}
	}

MetaIncluded:
	for _, meta := range res.Metas {
		if meta.Name[0] != '_' {
			for _, attr := range attrs {
				if attr == meta.FieldName || attr == meta.Name {
					continue MetaIncluded
				}
			}
			attrs = append(attrs, meta.Name)
		}
	}

	return attrs
}

func (res *Resource) SectionsList(values ...interface{}) (dest []*Section) {
	res.setSections(&dest, values...)
	return
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	}

	values := fc()
	var metas []*Meta
	for _, value := range values {
		metas = append(metas, value.(*Meta))
	}
	(*res.cachedMetas)[cacheKey] = metas
	return metas
}

// GetMetas get metas with give attrs
func (res *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = res.allAttrs()
	}
	var showSections, ignoredAttrs []string
	for _, attr := range attrs {
		if strings.HasPrefix(attr, "-") {
			ignoredAttrs = append(ignoredAttrs, strings.TrimLeft(attr, "-"))
		} else {
			showSections = append(showSections, attr)
		}
	}

	metas := []resource.Metaor{}

Attrs:
	for _, attr := range showSections {
		for _, a := range ignoredAttrs {
			if attr == a {
				continue Attrs
			}
		}

		var meta *Meta
		for _, m := range res.Metas {
			if m.GetName() == attr {
				meta = m
				break
			}
		}

		if meta == nil {
			meta = &Meta{Name: attr, baseResource: res, Resource: res}
			for _, primaryField := range res.PrimaryFields {
				if attr == primaryField.Name {
					meta.Type = "hidden_primary_key"
					break
				}
			}
			meta.updateMeta()
		}

		metas = append(metas, meta)
	}

	return metas
}

func (res *Resource) GetDefinedMeta(name string) *Meta {
	meta := res.MetasByName[name]
	if meta == nil {
		meta = res.MetasByFieldName[name]
	}
	return meta
}

// GetMeta get meta with name
func (res *Resource) GetMeta(name string, notUpdate ...bool) *Meta {
	fallbackMeta := res.MetasByName[name]

	if fallbackMeta == nil {
		fallbackMeta = res.MetasByFieldName[name]
	}

	if fallbackMeta == nil {
		if field, ok := res.FakeScope.FieldByName(name); ok {
			meta := &Meta{Name: field.Name, baseResource: res, Resource: res}
			if field.IsPrimaryKey {
				meta.Type = "hidden_primary_key"
			}
			if len(notUpdate) == 0 || !notUpdate[0] {
				meta.updateMeta()
			}
			res.MetasByName[meta.Name] = meta
			res.MetasByFieldName[name] = meta
			res.Metas = append(res.Metas, meta)
			return meta
		} else {
			parts := strings.Split(name, ".")
			if len(parts) > 1 {
				r := res
				for _, p := range parts[0 : len(parts)-1] {
					if r.Fragments != nil && r.Fragments.Get(p) != nil {
						r = r.Fragments.Get(p).Resource
						meta := r.GetMeta(parts[len(parts)-1])
						if meta != nil {
							meta = NewMetaFieldProxy(meta.Name, parts, meta)
							res.MetasByName[meta.Name] = meta
							res.MetasByFieldName[meta.Name] = meta
							res.Metas = append(res.Metas, meta)
							meta.updateMeta()
						}
						return meta
					} else {
						return nil
					}
				}

				meta := r.GetMeta(parts[len(parts)-1])
				if meta != nil {
					meta = NewMetaFieldProxy(name, parts, meta)
					res.MetasByName[name] = meta
					res.MetasByFieldName[name] = meta
					res.Metas = append(res.Metas, meta)
					meta.updateMeta()
					return meta
				}
			}
		}
	}

	return fallbackMeta
}

func DefaultPermission(action string, defaul ...roles.PermissionMode) roles.PermissionMode {
	switch action {
	case "index", "show":
		return roles.Read
	case "edit":
		return roles.Update
	}
	if len(defaul) == 0 {
		return roles.NONE
	}
	return defaul[0]
}

func (res *Resource) allowedSections(record interface{}, sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	var newSections []*Section
	for _, section := range sections {
		newSection := &Section{Resource: section.Resource, Title: section.Title}
		var editableRows [][]string
		for _, row := range section.Rows {
			var editableColumns []string
			for _, column := range row {
				meta := section.Resource.GetMeta(column)
				if meta != nil {
					if meta.Enabled != nil && !meta.Enabled(record, context, meta) {
						continue
					}

					for _, role := range roles {
						if meta.HasPermission(role, context.Context) {
							editableColumns = append(editableColumns, column)
							break
						}
					}
				}
			}
			if len(editableColumns) > 0 {
				editableRows = append(editableRows, editableColumns)
			}
		}

		if len(editableRows) > 0 {
			newSection.Rows = editableRows
			newSections = append(newSections, newSection)
		}
	}
	return newSections
}

func (res *Resource) MetasFromLayoutContext(layout string, context *Context, value interface{}, roles ...roles.PermissionMode) (metas []*Meta, names []*resource.MetaName) {
	if len(roles) == 0 {
		defaultRole := DefaultPermission(layout)
		roles = append(roles, defaultRole)
	}
	l := res.GetLayout(layout).(*Layout)
	if l != nil {
		if l.MetasFunc != nil {
			metas, names = l.MetasFunc(res, context, value, roles...)
		} else if l.MetaNamesFunc != nil {
			namess := l.MetaNamesFunc(res, context, value, roles...)
			if len(namess) > 0 {
				metas = res.ConvertSectionToMetas(res.allowedSections(value, res.generateSections(namess), context, roles...))
			}
		} else if len(l.Metas) > 0 {
			for _, metaName := range l.Metas {
				metas = append(metas, res.MetasByName[metaName])
			}

			names = l.MetaNames
		}

		if len(metas) > 0 && len(names) == 0 {
			names = make([]*resource.MetaName, len(metas), len(metas))

			if l.MetaAliases == nil {
				for i, meta := range metas {
					names[i] = meta.Namer()
				}
			} else {
				for i, meta := range metas {
					if alias, ok := l.MetaAliases[meta.Name]; ok {
						names[i] = alias
					} else {
						names[i] = meta.Namer()
					}
				}
			}
		}
	}
	return
}

func (res *Resource) BasicValue(record interface{}) resource.BasicValue {
	metaId, metaLabel, metaIcon := res.MetasByName[BASIC_META_ID], res.MetasByName[BASIC_META_LABEL], res.MetasByName[BASIC_META_ICON]
	return &resource.Basic{
		metaId.GetFormattedValuer()(record, nil).(string),
		metaLabel.GetFormattedValuer()(record, nil).(string),
		metaIcon.GetFormattedValuer()(record, nil).(string),
	}
}

func (res *Resource) MountTo(param string) *Resource {
	config := &(*res.Config)
	if config.Sub != nil {
		config.Sub = &(*config.Sub)
	}
	nmp := utils.NamifyString(param)
	config.Name += nmp
	config.Param = param
	config.ID += nmp
	config.NotMount = false
	config.Invisible = true
	return res.admin.AddResource(res.Value, config)
}

func (res *Resource) GetDefaultRouterActions(object bool) []string {
	if res.Config.Singleton {
		return []string{A_READ, A_UPDATE}
	}
	r := []string{A_CREATE, A_READ, A_UPDATE, A_DELETE}
	if !object {
		r = append(r, A_INDEX)
	}
	return r
}

func DefaultRouterPathAndMethod(action string, form, object, singleton bool) (pth, method string) {
	switch strings.ToLower(action) {
	case A_CREATE:
		if form {
			return P_NEW_FORM, "GET"
		}
		return P_NEW, "POST"
	case A_UPDATE:
		if singleton || object {
			if form {
				return P_OBJ_UPDATE_FORM, "GET"
			}
			return P_OBJ_UPDATE, "PUT"
		} else {
			if form {
				return "/{id}" + P_OBJ_READ_FORM, "GET"
			}
			return "/{id}" + P_OBJ_UPDATE, "PUT"
		}
	case A_READ:
		if singleton || object {
			return P_OBJ_READ, "GET"
		}
		return "/{id}" + P_OBJ_READ_FORM, "GET"
	case A_INDEX:
		if !object && !singleton {
			return "/", "GET"
		}
	case A_DELETE:
		if !singleton {
			if object {
				return P_OBJ_DELETE, "DELETE"
			}
			return "/{id}" + P_OBJ_DELETE, "DELETE"
		}
	}
	return "", ""
}

func (res *Resource) RegisterDefaultRouters(actions ...string) {
	if len(actions) == 0 {
		actions = []string{"create", "update", "read", "delete"}
	}

	var (
		adminController = &Controller{Admin: res.GetAdmin()}
	)

	if res.AdminController == nil {
		res.AdminController = adminController
	}

	for _, action := range actions {
		switch strings.ToLower(action) {
		case "create":
			if !res.Config.Singleton {
				// New
				res.Router.Get(P_NEW_FORM, NewHandler(adminController.New, &RouteConfig{PermissionMode: roles.Create, Resource: res}))
			}

			res.Router.Api(func(router *route.Mux) {
				// Create
				router.Post(P_NEW, NewHandler(adminController.Create, &RouteConfig{PermissionMode: roles.Create, Resource: res}))
			})
		case "update":
			if res.Config.Singleton {
				// Edit
				res.Router.Get(P_SINGLETON_UPDATE_FORM, NewHandler(adminController.Edit, &RouteConfig{PermissionMode: roles.Update, Resource: res}))
				res.Router.Api(func(router *route.Mux) {
					// Update
					router.Put(P_SINGLETON_UPDATE, NewHandler(adminController.Update, &RouteConfig{PermissionMode: roles.Update, Resource: res}))
				})
			} else {
				// Edit
				res.ObjectRouter.Get(P_OBJ_UPDATE_FORM, NewHandler(adminController.Edit, &RouteConfig{PermissionMode: roles.Update, Resource: res}))

				res.ObjectRouter.Api(func(router *route.Mux) {
					update := NewHandler(adminController.Update, &RouteConfig{PermissionMode: roles.Update, Resource: res})
					// Update
					router.Put(P_OBJ_UPDATE, update)
					router.Post(P_OBJ_UPDATE, update)
				})
			}
		case "read":
			res.Router.Api(func(router *route.Mux) {
				if res.Config.Singleton {
					// Show
					router.Get(P_SINGLETON_READ_FORM, NewHandler(adminController.Show, &RouteConfig{PermissionMode: roles.Read, Resource: res}))
				} else {
					// Index
					router.Get(P_INDEX, NewHandler(adminController.Index, &RouteConfig{PermissionMode: roles.Read, Resource: res}))

				}
			})
			res.ObjectRouter.Api(func(router *route.Mux) {
				// Show
				router.Get(P_OBJ_READ_FORM, NewHandler(adminController.Show, &RouteConfig{PermissionMode: roles.Read, Resource: res}))
			})
		case "delete":
			if !res.Config.Singleton {
				// Delete
				res.ObjectRouter.Delete(P_OBJ_DELETE, NewHandler(adminController.Delete, &RouteConfig{PermissionMode: roles.Delete, Resource: res}))
			}
		}
	}
}

func (res *Resource) CreateMenu(plural bool) *Menu {
	menuName := res.Name

	if plural {
		menuName = inflection.Plural(menuName)
	}

	menu := &Menu{
		Name:         menuName,
		Label:        res.GetLabelKey(plural),
		Permissioner: res,
		Priority:     res.Config.Priority,
		Ancestors:    res.Config.Menu,
		RelativePath: res.GetIndexURI(),
		Enabled:      res.Config.MenuEnabled,
		Resource:     res,
	}

	if res.ParentResource != nil {
		menu.MakeLink = func(context *Context, args ...interface{}) string {
			var parentKeys []string
			for _, arg := range args {
				switch t := arg.(type) {
				case string:
					if t != "" {
						parentKeys = append(parentKeys, t)
					}
				case []string:
					parentKeys = append(parentKeys, t...)
				}
			}
			if len(parentKeys) == 0 {
				return res.GetContextIndexURI(context.Context)
			}
			return res.GetContextIndexURI(context.Context, parentKeys...)
		}
	}

	return menu
}

func (res *Resource) GetDefaultMenu() *Menu {
	if res.DefaultMenu == nil {
		res.DefaultMenu = res.CreateMenu(!res.Config.Singleton)
	}
	return res.DefaultMenu
}

func (res *Resource) GetIndexLink(context *core.Context, args ...interface{}) string {
	return res.GetLink(nil, context, args...)
}

func (res *Resource) GetLink(record interface{}, context *core.Context, args ...interface{}) string {
	var parentKeys []string
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			if t != "" {
				parentKeys = append(parentKeys, t)
			}
		case []string:
			parentKeys = append(parentKeys, t...)
		}
	}
	if record == nil {
		return res.GetContextIndexURI(context, parentKeys...)
	}
	uri := res.GetRecordURI(record, parentKeys...)
	return context.GenURL(uri)
}

func (res *Resource) configure() {
	modelType := utils.ModelType(res.Value)

	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceInterface); ok {
				injector.ConfigureQorResource(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceInterface); ok {
		injector.ConfigureQorResource(res)
	}

	typ := reflect.TypeOf(res.Value)

	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	_, softDelete := typ.FieldByName("DeletedAt")

	res.Action(&Action{
		Name:   "Delete",
		Method: "DELETE",
		Type:   ActionDanger,
		URL: func(record interface{}, context *Context, args ...interface{}) string {
			return res.GetLink(record, context.Context, args...)
		},
		Permission: res.Config.Permission,
		Modes:      []string{"menu_item"},
		Visible: func(record interface{}, context *Context) bool {
			if softDelete {
				typ := reflect.ValueOf(record)

				for typ.Kind() == reflect.Ptr {
					typ = typ.Elem()
				}

				f := typ.FieldByName("DeletedAt")
				if f.IsValid() {
					v := f.Interface()
					if t, ok := v.(time.Time); ok {
						return t.IsZero()
					} else if t, ok := v.(*time.Time); ok {
						return t == nil || t.IsZero()
					}
				}
				return false
			}
			return true
		},
	})
}

// GetResources get defined resources from admin
func (res *Resource) GetResources() (resources []*Resource) {
	for _, r := range res.Resources {
		resources = append(resources, r)
	}
	return
}

func (res *Resource) WalkResources(f func(res *Resource) bool) bool {
	for _, r := range res.Resources {
		if !f(r) {
			break
		}
		if !r.WalkResources(f) {
			break
		}
	}
	return true
}

// GetResourceByID get resource with name
func (res *Resource) GetResourceByID(id string) (resource *Resource) {
	parts := strings.SplitN(id, ".", 2)
	r := res.Resources[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByID(parts[1])
	}
}

// GetResourceByParam get resource with name
func (res *Resource) GetResourceByParam(param string) (resource *Resource) {
	parts := strings.SplitN(param, ".", 2)
	r := res.ResourcesByParam[parts[0]]
	if r == nil || len(parts) == 1 {
		return r
	} else {
		return r.GetResourceByParam(parts[1])
	}
}

func (res *Resource) GetParentResourceByID(id string) *Resource {
	for _, p := range res.Parents {
		if p.ID == id {
			return p
		}
	}
	return res.admin.GetResourceByID(id)
}

func (res *Resource) GetOrParentResourceByID(id string) *Resource {
	r := res.GetResourceByID(id)
	if r == nil {
		r = res.GetParentResourceByID(id)
	}
	return r
}

func (res *Resource) SubResources() (items []*Resource) {
	for _, r := range res.Resources {
		if !r.Config.Invisible {
			items = append(items, r)
		}
	}
	return
}

func (res *Resource) ReferencedRecord(record interface{}) interface{} {
	return nil
}

func (res *Resource) CrudScheme(ctx *core.Context, scheme interface{}) *resource.CRUD {
	s := res.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = res.GetScheme(st)
	default:
		if scheme != nil {
			s = scheme.(*Scheme)
		}
	}
	return res.Crud(ctx).Dispatcher(s.EventDispatcher)
}

func (res *Resource) CrudSchemeDB(db *aorm.DB, scheme interface{}) *resource.CRUD {
	s := res.Scheme
	switch st := scheme.(type) {
	case string:
		s, _ = res.GetScheme(st)
	default:
		if scheme != nil {
			s = scheme.(*Scheme)
		}
	}
	return res.CrudDB(db).Dispatcher(s.EventDispatcher)
}

func (res *Resource) Crud(ctx *core.Context) *resource.CRUD {
	return resource.NewCrud(res, ctx)
}

func (res *Resource) CrudDB(db *aorm.DB) *resource.CRUD {
	return res.Crud(&core.Context{DB: db})
}

func (res *Resource) SetParentResource(parent *Resource, fieldName string) {
	res.Resource.SetParent(parent, fieldName)
	res.ParentResource = parent
}

func (res *Resource) RegisterScheme(name string, setup ...func(s *Scheme)) *Scheme {
	f := func() *Scheme {
		return res.Scheme.AddChild(name, setup...)
	}
	if res.registered {
		return f()
	}
	res.afterRegister = append(res.afterRegister, func() {
		f()
	})
	return nil
}

func (res *Resource) triggerSchemeAdded(s *Scheme) {
	s.Resource.Trigger(&SchemeEvent{edis.NewEvent(E_SCHEME_ADDED), s})
}

func (res *Resource) HasScheme(name string) bool {
	_, ok := res.GetScheme(name)
	return ok
}

func (res *Resource) DefaultFilter(fns ...func(context *core.Context, db *aorm.DB) *aorm.DB) {
	res.Scheme.DefaultFilter(fns...)
}

func (res *Resource) AddFragmentConfig(value fragment.FragmentModelInterface, cfg *FragmentConfig) *Resource {
	if _, ok := res.Value.(fragment.FragmentedModelInterface); !ok {
		panic(NotFragmentableError)
	}

	if cfg == nil {
		cfg = &FragmentConfig{}
	}
	if cfg.Config == nil {
		cfg.Config = &Config{}
	}
	if cfg.Config.Sub == nil {
		cfg.Config.Sub = &SubConfig{}
	}

	cfg.Config.Singleton = true
	cfg.Config.Sub.Parent = res
	cfg.Config.DisableFormID = true
	cfg.Config.Sub.ParentFieldName = "ID"

	_, isForm := value.(fragment.FormFragmentModelInterface)

	setup := cfg.Config.Setup

	cfg.Config.Setup = func(fragRes *Resource) {
		fragRes.SetMeta(&Meta{Name: "ID", Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
			return false
		}}, true)
		if isForm {
			meta := &Meta{
				Name: "FragmentEnabled",
			}
			if cfg.Is {
				meta.SkipDefaultLabel = true
				if cfg.IsLabel != "" {
					meta.Label = cfg.IsLabel
				} else {
					meta.Label = fragRes.SingularLabelKey()
				}
				if meta.Label[0] == '.' {
					meta.Label = fragRes.I18nPrefix + meta.Label
				}
				meta.Enabled = func(recorde interface{}, context *Context, meta *Meta) bool {
					if context.Type == SHOW {
						return false
					}
					return true
				}
			} else {
				meta.SkipDefaultLabel = true
				meta.Label = fragRes.SingularLabelKey()
			}
			fragRes.SetMeta(meta, true)
			res.Fragments.AddForm(fragRes, cfg)
		} else {
			res.Fragments.Add(fragRes, cfg)
		}

		fragRes.Fragment.Build()

		if setup != nil {
			setup(fragRes)
		}
	}

	if !isForm {
		cfg.Config.Invisible = true
	} else {
		if cfg.Is || cfg.NotInline {
			cfg.Config.MenuEnabled = func(menu *Menu, ctx *Context) bool {
				if r, ok := ctx.Result.(fragment.FragmentedModelInterface); ok {
					if f := r.GetFormFragment(menu.Resource.Fragment.ID); f != nil {
						return f.Enabled() && (menu.Resource.Fragment.Config.Is || menu.Resource.Fragment.Config.NotInline)
					}
				}
				return false
			}
		}
	}

	fragRes := res.AddResourceConfig(value, cfg.Config)

	if len(res.Fragments.Fragments) == 1 {
		res.GetAdmin().OnDone(func(e *AdminEvent) {
			res.Fragments.Build()
		})
		res.Filter(&Filter{
			Name: "Only",
			Available: func(context *core.Context) (ok bool) {
				for _, f := range res.Fragments.Fragments {
					if f.Config.Is {
						return true
					}
				}
				return false
			},
			//Label: I18NGROUP + ".actions.fragment_is_filter",
			Handler: func(db *aorm.DB, argument *FilterArgument) *aorm.DB {
				if v := argument.Value.Get("Value").Value; v != nil {
					if vs, ok := v.([]string); ok && len(vs) > 0 {
						id := vs[0]
						if f := argument.Resource.Fragments.Get(id); f != nil {
							return f.Filter(db)
						}
					}
				}
				return db
			},
			Config: &SelectOneConfig{
				Collection: func(context *Context) (results [][]string) {
					for id, f := range res.Fragments.Fragments {
						if f.Config.Is {
							results = append(results, []string{id, context.Ts(f.Resource.PluralLabelKey(), f.Resource.PluralName)})
						}
					}
					return
				},
			},
		})
		res.OnDBActionE(func(e *resource.DBEvent) (err error) {
			context := e.Context
			if v := context.Data().Get("skip.fragments"); v == nil {
				r := e.Result().(fragment.FragmentedModelInterface)
				for id, fr := range r.GetFragments() {
					fragRes := res.Fragments.Get(id).Resource
					if fr.GetID() == "" {
						fr.SetID(r.GetID())
					}
					if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
						return errwrap.Wrap(err, "Fragment "+id)
					}
				}
				for id, fr := range r.GetFormFragments() {
					fragRes := res.Fragments.Get(id).Resource
					if fr.GetID() == "" {
						fr.SetID(r.GetID())
					}
					if err = fragRes.Crud(e.OriginalContext()).Update(fr); err != nil {
						return errwrap.Wrap(err, "Fragment "+id)
					}
				}
			}
			return nil
		}, resource.E_DB_ACTION_SAVE.After())

		res.Scheme.OnDBActionE(func(e *resource.DBEvent) (err error) {
			context := e.Context
			if v := context.Data().Get("skip.fragments"); v == nil {
				DB := context.DB
				fields, query := res.Fragments.Fields(), res.Fragments.Query()
				DB = DB.ExtraSelectFieldsSetter(
					PKG+".fragments",
					func(result interface{}, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
						res.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
					}, fields, query)
				DB = res.Fragments.JoinLeft(DB)
				context.SetDB(DB)
			}
			return nil
		}, resource.BEFORE|resource.E_DB_ACTION_FIND_MANY|resource.E_DB_ACTION_FIND_ONE)

		/*res.OnDBActionE(func(e *resource.DBEvent) (err error) {
			context := e.Context
			if v := context.Data().Get("skip.fragments"); v == nil {
				DB := context.DB
				fields, query := res.Fragments.Fields(), res.Fragments.Query()
				DB = DB.ExtraSelectFieldsSetter(
					PKG+".fragments",
					func(result interface{}, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
						res.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
					}, fields, query)
				DB = res.Fragments.JoinLeft(DB)
				context.SetDB(DB)
			}
			return nil
		}, resource.BEFORE|resource.E_DB_ACTION_FIND_MANY|resource.E_DB_ACTION_FIND_ONE)*/

		res.FakeScope.GetModelStruct().BeforeRelatedCallback(func(fromScope *aorm.Scope, toScope *aorm.Scope, DB *aorm.DB, fromField *aorm.Field) *aorm.DB {
			fields, query := res.Fragments.Fields(), res.Fragments.Query()
			DB = DB.ExtraSelectFieldsSetter(
				PKG+".fragments",
				func(result interface{}, values []interface{}, set func(result interface{}, low, hight int) interface{}) {
					res.Fragments.ExtraFieldsScan(result.(fragment.FragmentedModelInterface), values, set)
				}, fields, query)
			DB = res.Fragments.JoinLeft(DB)
			return DB
		})
	}

	metaName := fragRes.Fragment.ID

	if !isForm {
		fieldsNames := fragRes.Fragment.FieldsNames()
		for _, field := range fieldsNames {
			meta := NewMetaProxy(field, fragRes.Meta(&Meta{Name: field}), func(meta *Meta, recorde interface{}) interface{} {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				return fragmentedRecorde.GetFragment(frag.ID)
			})
			meta.Fragment = fragRes.Fragment
			meta.Resource = res
			meta.NewValuer(func(meta *Meta, old MetaValuer, recorde interface{}, context *core.Context) interface{} {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				value := frag.GetOrNew(fragmentedRecorde, context)
				return meta.ProxyTo.GetValuer()(value, context)
			})
			meta.NewSetter(func(meta *Meta, old MetaSetter, recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				fragmentedRecorde := recorde.(fragment.FragmentedModelInterface)
				frag := meta.Fragment
				value := frag.GetOrNew(fragmentedRecorde, context)
				return meta.ProxyTo.GetSetter()(value, metaValue, context)
			})
			meta = res.Meta(meta)
		}

		fieldsNamesInterface := helpers.StringsToInterfaces(fieldsNames)
		res.EditAttrs(append([]interface{}{res.EditAttrs()}, fieldsNamesInterface...)...)
		res.ShowAttrs(append([]interface{}{res.ShowAttrs()}, fieldsNamesInterface...)...)
	} else {
		res.Meta(&Meta{
			Name: metaName,
			Type: "fragment",
			Enabled: func(recorde interface{}, context *Context, meta *Meta) bool {
				if recorde != nil {
					return fragRes.Fragment.Enabled(recorde.(fragment.FragmentedModelInterface), context.Context)
				}
				return false
			},
			ContextMetas: func(record interface{}, ctx *Context) []*Meta {
				return fragRes.ConvertSectionToMetas(fragRes.ContextSections(ctx, record))
			},
			Setter: func(recorde interface{}, metaValue *resource.MetaValue, context *core.Context) error {
				if _, ok := res.Fragments.Fragments[metaValue.Name]; !ok {
					return nil
				}
				value := fragRes.Fragment.FormGetOrNew(recorde.(fragment.FragmentedModelInterface), context)
				err := resource.DecodeToResource(fragRes, value, metaValue.MetaValues, context).Start()
				return err
			},
			Valuer: func(recorde interface{}, context *core.Context) interface{} {
				r := recorde.(fragment.FragmentedModelInterface)
				value := r.GetFormFragment(fragRes.Fragment.ID)
				isNil := value == nil
				if isNil {
					value = fragRes.NewStruct(context.Site).(fragment.FormFragmentModelInterface)
				}
				return &FormFragmentRecordState{
					fragRes.Fragment,
					fragRes.Fragment.Enabled(r, context),
					value,
					isNil,
				}
			},
		})

		var hasEditMeta bool
		if len(res.editSections) > 0 {
		root:
			for _, sec := range res.editSections {
				for _, row := range sec.Rows {
					for _, col := range row {
						if col == metaName {
							hasEditMeta = true
							break root
						}
					}
				}
			}
		}

		if !hasEditMeta {
			res.EditAttrs(res.EditAttrs(), metaName)
		}

		if len(res.showSections) == 0 {
			res.ShowAttrs(res.EditAttrs(), metaName)
		} else {
			var hasShowMeta bool
		root2:
			for _, sec := range res.showSections {
				for _, row := range sec.Rows {
					for _, col := range row {
						if col == metaName {
							hasShowMeta = true
							break root2
						}
					}
				}
			}

			if !hasShowMeta {
				res.ShowAttrs(res.ShowAttrs(), metaName)
			}
		}
	}

	return fragRes
}

func (res *Resource) GetAdminLayout(name string, defaul ...string) *Layout {
	return res.GetLayout(name, defaul...).(*Layout)
}

func (res *Resource) AddFragment(value fragment.FragmentModelInterface) *Resource {
	return res.AddFragmentConfig(value, &FragmentConfig{})
}
