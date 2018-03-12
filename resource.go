package admin

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/moisespsena/template/html/template"
)

type SubResourceConfig struct {
	FieldName     string
	LabelPlural   string
	LabelSingular string
	IconSingular  string
	IconPlural    string
	Invisible     bool
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

	var path []string
	for i, r := 0, res.Resource.ParentResource; r != nil; i++ {
		path = append(path, parentParams[i], r.ToParam())
		r = r.ParentResource
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	path = append(path, res.Resource.ToParam())
	menu.RelativePath = strings.Join(path, "/")

	return menu
}

func (res *SubResource) CreateDefaultMenu(parentParams ...string) *Menu {
	return res.CreateMenu(!res.Resource.Config.Singleton, parentParams...)
}

// Resource is the most important thing for qor admin, every model is defined as a resource, qor admin will genetate management interface based on its definition
type Resource struct {
	*resource.Resource
	Config         *Config
	Metas          []*Meta
	Actions        []*Action
	SearchHandler  func(keyword string, context *qor.Context) *gorm.DB
	ParentResource *Resource

	admin           *Admin
	params          string
	mounted         bool
	scopes          []*Scope
	filters         map[string]*Filter
	sortableAttrs   *[]string
	indexSections   []*Section
	newSections     []*Section
	editSections    []*Section
	showSections    []*Section
	isSetShowAttrs  bool
	cachedMetas     *map[string][]*Meta
	customSections  *map[string]*[]*Section
	AdminController *Controller
	SubResources    map[string]*SubResource
}

func (res *Resource) GetLabelKey(plural bool) string {
	r := res.I18nPrefix + ".label"
	if plural {
		r += "~p"
	} else {
		r += "~s"
	}
	return r
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
func (res *Resource) Meta(meta *Meta) *Meta {
	if oldMeta := res.GetMeta(meta.Name); oldMeta != nil {
		if meta.Type != "" {
			oldMeta.Type = meta.Type
			oldMeta.Config = nil
		}

		if meta.TypeHander != nil {
			oldMeta.TypeHander = meta.TypeHander
		}

		if meta.Enabled != nil {
			oldMeta.Enabled = meta.Enabled
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

		meta = oldMeta
	} else {
		res.Metas = append(res.Metas, meta)
		meta.baseResource = res
	}

	meta.updateMeta()
	return meta
}

// GetAdmin get admin from resource
func (res Resource) GetAdmin() *Admin {
	return res.admin
}

// GetURL
func (res Resource) GetPrefix(context *Context, parentkeys ... string) string {
	var params string
	if res.ParentResource != nil {
		if context == nil {
			params = path.Join(res.ParentResource.GetPrefix(nil, parentkeys[0:len(parentkeys)-1]...), parentkeys[len(parentkeys)-1])
		} else {
			params = path.Join(res.ParentResource.GetPrefix(context), res.ParentResource.GetPrimaryValue(context.Request))
		}
	}
	return params
}

// GetURL
func (res Resource) GetURI(context *Context, parentkeys ... string) string {
	return context.GenURL(path.Join(res.GetPrefix(context, parentkeys...), res.ToParam()))
}

// GetURL
func (res Resource) GetURIForKey(context *Context, key string, parentkeys ... string) string {
	if key == "" {
		key = res.GetPrimaryValue(context.Request)
	}
	return path.Join(res.GetURI(context, parentkeys...), key)
}

// GetPrimaryValue get priamry value from request
func (res Resource) GetPrimaryValue(request *http.Request) string {
	if request != nil {
		return request.URL.Query().Get(res.ParamIDName())
	}
	return ""
}

// ParamIDName return param name for primary key like :product_id
func (res Resource) ParamIDName() string {
	return fmt.Sprintf(":%v_id", inflection.Singular(utils.ToParamString(res.Name)))
}

// ToParam used as urls to register routes for resource
func (res *Resource) ToParam() string {
	if res.params == "" {
		if res.Config.Param != "" {
			res.params = res.Config.Param
		} else if value, ok := res.Value.(interface {
			ToParam() string
		}); ok {
			res.params = value.ToParam()
		} else {
			if res.Config.Singleton {
				res.params = utils.ToParamString(res.Name)
			}
			res.params = utils.ToParamString(inflection.Plural(res.Name))
		}
	}
	return res.params
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

		// Config Admin Theme
		for _, pth := range themeInterface.GetViewPaths() {
			res.GetAdmin().RegisterViewPath(pth)
		}
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
func (res *Resource) NewResource(value interface{}, config ...*Config) *Resource {
	subRes := res.GetAdmin().newResource(value, config...)
	subRes.ParentResource = res
	subRes.configure()
	return subRes
}

// AddSubResource register sub-resource
func (res *Resource) AddSubResource(resourceConfig *SubResourceConfig, config ...*Config) (subRes *Resource, err error) {
	var (
		admin = res.GetAdmin()
		scope = &gorm.Scope{Value: res.Value}
	)

	if field, ok := scope.FieldByName(resourceConfig.FieldName); ok && field.Relationship != nil {
		modelType := utils.ModelType(reflect.New(field.Struct.Type).Interface())

		var cfg *Config
		if len(config) > 0 {
			cfg = config[0]
		} else {
			cfg = &Config{}
		}

		if cfg.Name == "" {
			cfg.Name = resourceConfig.FieldName
		}

		if cfg.Param == "" {
			cfg.Param = utils.ToParamString(cfg.Name)
		}

		subRes = admin.NewResource(reflect.New(modelType).Interface(), cfg)
		localParam := subRes.ToParam()
		subRes.setupParentResource(field.StructField.Name, res)
		subRes.RegisterDefaultRouters()

		if res.SubResources == nil {
			res.SubResources = make(map[string]*SubResource)
		}

		sres := &SubResource{subRes, resourceConfig}

		res.SubResources[localParam] = sres

		return
	}

	err = errors.New("invalid sub resource")
	return
}

func (res *Resource) setupParentResource(fieldName string, parent *Resource) {
	res.ParentResource = parent

	findOneHandler := res.FindOneHandler
	res.FindOneHandler = func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) (err error) {
		if metaValues != nil {
			return findOneHandler(value, metaValues, context)
		}

		if primaryKey := res.GetPrimaryValue(context.Request); primaryKey != "" {
			clone := context.Clone()
			parentValue := parent.NewStruct(context.Site)
			if err = parent.FindOneHandler(parentValue, nil, clone); err == nil {
				primaryQuerySQL, primaryParams := res.ToPrimaryQueryParams(primaryKey, context)
				err = context.DB.Model(parentValue).Where(primaryQuerySQL, primaryParams...).Related(value).Error
			}
		}
		return
	}

	res.FindManyHandler = func(value interface{}, context *qor.Context) error {
		var (
			err         error
			clone       = context.Clone()
			parentValue = parent.NewStruct(context.Site)
		)

		if err = parent.FindOneHandler(parentValue, nil, clone); err == nil {
			parent.FindOneHandler(parentValue, nil, clone)
			return context.DB.Model(parentValue).Related(value).Error
		}
		return err
	}

	res.SaveHandler = func(value interface{}, context *qor.Context) error {
		var (
			err         error
			clone       = context.Clone()
			parentValue = parent.NewStruct(context.Site)
		)

		if err = parent.FindOneHandler(parentValue, nil, clone); err == nil {
			parent.FindOneHandler(parentValue, nil, clone)
			return context.DB.Model(parentValue).Association(fieldName).Append(value).Error
		}
		return err
	}

	res.DeleteHandler = func(value interface{}, context *qor.Context) (err error) {
		var clone = context.Clone()
		var parentValue = parent.NewStruct(context.Site)
		if primaryKey := res.GetPrimaryValue(context.Request); primaryKey != "" {
			primaryQuerySQL, primaryParams := res.ToPrimaryQueryParams(primaryKey, context)
			if err = context.DB.Where(primaryQuerySQL, primaryParams...).First(value).Error; err == nil {
				if err = parent.FindOneHandler(parentValue, nil, clone); err == nil {
					parent.FindOneHandler(parentValue, nil, clone)
					return context.DB.Model(parentValue).Association(fieldName).Delete(value).Error
				}
			}
		}
		return
	}
}

// Decode decode context into a value
func (res *Resource) Decode(context *qor.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &gorm.Scope{Value: res.Value}

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
		for _, attr := range attrs {
			if attr == meta.FieldName || attr == meta.Name {
				continue MetaIncluded
			}
		}
		attrs = append(attrs, meta.Name)
	}

	return attrs
}

func (res *Resource) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return res.allAttrs()
	}

	var onlyExcludeAttrs = true
	for _, attr := range attrs {
		if !strings.HasPrefix(attr, "-") {
			onlyExcludeAttrs = false
			break
		}
	}

	if onlyExcludeAttrs {
		return append(res.allAttrs(), attrs...)
	}
	return attrs
}

func (res *Resource) GetCustomAttrs(name string) ([]*Section, bool) {
	if res.customSections == nil {
		return nil, false
	}
	sections, ok := (*res.customSections)[name]
	if ok {
		return *sections, ok
	} else {
		return nil, false
	}
}

// IndexAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (res *Resource) CustomAttrs(name string, values ...interface{}) []*Section {
	if res.customSections == nil {
		res.customSections = &map[string]*[]*Section{}
	}

	sections := &[]*Section{}
	res.setSections(sections, values...)
	(*res.customSections)[name] = sections

	return *sections
}

// IndexAttrs set attributes will be shown in the index page
//     // show given attributes in the index page
//     order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
//     // show all attributes except `State` in the index page
//     order.IndexAttrs("-State")
func (res *Resource) IndexAttrs(values ...interface{}) []*Section {
	res.setSections(&res.indexSections, values...)
	res.SearchAttrs()
	return res.indexSections
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
func (res *Resource) NewAttrs(values ...interface{}) []*Section {
	res.setSections(&res.newSections, values...)
	return res.newSections
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
func (res *Resource) EditAttrs(values ...interface{}) []*Section {
	res.setSections(&res.editSections, values...)
	return res.editSections
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
func (res *Resource) ShowAttrs(values ...interface{}) []*Section {
	if len(values) > 0 {
		if values[len(values)-1] == false {
			values = values[:len(values)-1]
		} else {
			res.isSetShowAttrs = true
		}
	}
	res.setSections(&res.showSections, values...)
	return res.showSections
}

// SortableAttrs set sortable attributes, sortable attributes could be click to order in qor table
func (res *Resource) SortableAttrs(columns ...string) []string {
	if len(columns) != 0 || res.sortableAttrs == nil {
		if len(columns) == 0 {
			columns = res.ConvertSectionToStrings(res.indexSections)
		}
		res.sortableAttrs = &[]string{}
		scope := qor.FakeDB.NewScope(res.Value)
		for _, column := range columns {
			if field, ok := scope.FieldByName(column); ok && field.DBName != "" {
				attrs := append(*res.sortableAttrs, column)
				res.sortableAttrs = &attrs
			}
		}
	}
	return *res.sortableAttrs
}

// SearchAttrs set search attributes, when search resources, will use those columns to search
//     // Search products with its name, code, category's name, brand's name
//	   product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
func (res *Resource) SearchAttrs(columns ...string) []string {
	if len(columns) != 0 || res.SearchHandler == nil {
		if len(columns) == 0 {
			columns = res.ConvertSectionToStrings(res.indexSections)
		}

		if len(columns) > 0 {
			res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
				var filterFields []filterField
				for _, column := range columns {
					filterFields = append(filterFields, filterField{FieldName: column})
				}
				return filterResourceByFields(res, filterFields, keyword, context.DB, context)
			}
		}
	}

	return columns
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
			meta = &Meta{Name: attr, baseResource: res}
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

// GetMeta get meta with name
func (res *Resource) GetMeta(name string) *Meta {
	var fallbackMeta *Meta

	for _, meta := range res.Metas {
		if meta.Name == name {
			return meta
		}

		if meta.GetFieldName() == name {
			fallbackMeta = meta
		}
	}

	if fallbackMeta == nil {
		if field, ok := qor.FakeDB.NewScope(res.Value).FieldByName(name); ok {
			meta := &Meta{Name: name, baseResource: res}
			if field.IsPrimaryKey {
				meta.Type = "hidden_primary_key"
			}
			meta.updateMeta()
			res.Metas = append(res.Metas, meta)
			return meta
		}
	}

	return fallbackMeta
}

func (res *Resource) allowedSections(sections []*Section, context *Context, roles ...roles.PermissionMode) []*Section {
	var newSections []*Section
	for _, section := range sections {
		newSection := Section{Resource: section.Resource, Title: section.Title}
		var editableRows [][]string
		for _, row := range section.Rows {
			var editableColumns []string
			for _, column := range row {
				meta := res.GetMeta(column)
				if meta != nil {
					if meta.Enabled != nil && !meta.Enabled(context, meta) {
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
			newSections = append(newSections, &newSection)
		}
	}
	return newSections
}

func (res *Resource) RegisterDefaultRouters() {
	res.admin.RegisterResourceRouters(res, "create", "update", "read", "delete")
}

func (res *Resource) CreateMenu(plural bool) *Menu {
	menuName := res.Name

	if plural {
		menuName = inflection.Plural(menuName)
	}

	return &Menu{
		Name:         menuName,
		Label:        res.GetLabelKey(plural),
		Permissioner: res,
		Priority:     res.Config.Priority,
		Ancestors:    res.Config.Menu,
		RelativePath: res.ToParam(),
	}
}

func (res *Resource) CreateDefaultMenu() *Menu {
	return res.CreateMenu(!res.Config.Singleton)
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
		URL: func(record interface{}, context *Context) string {
			return context.URLFor(record, res)
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

func (res *Resource) VisibleChildren() (r []*SubResource) {
	for _, sub := range res.SubResources {
		if !sub.Config.Invisible {
			r = append(r, sub)
		}
	}
	return
}
