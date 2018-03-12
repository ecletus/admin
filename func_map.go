package admin

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/moisespsena/template/html/template"
	"github.com/moisespsena/template/funcs"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/session"
)

var TemplateExecutorMetaValue *template.Executor

func init() {
	t, err := template.New("qor:admin:meta_value").Parse("{{.Value}}")
	if err != nil {
		panic(err)
	}
	TemplateExecutorMetaValue = t.CreateExecutor()
}

// NewResourceContext new resource context
func (context *Context) NewResourceContext(name ...interface{}) *Context {
	clone := &Context{Context: context.Context.Clone(), Admin: context.Admin, Result: context.Result, Action: context.Action}
	if len(name) > 0 {
		if str, ok := name[0].(string); ok {
			clone.setResource(context.Admin.GetResource(str))
		} else if res, ok := name[0].(*Resource); ok {
			clone.setResource(res)
		}
	} else {
		clone.setResource(context.Resource)
	}
	return clone
}

func (context *Context) primaryKeyOf(value interface{}) interface{} {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		return fmt.Sprint(scope.PrimaryKeyValue())
	}
	return fmt.Sprint(value)
}

func (context *Context) uniqueKeyOf(value interface{}) interface{} {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		var primaryValues []string
		for _, primaryField := range scope.PrimaryFields() {
			primaryValues = append(primaryValues, fmt.Sprint(primaryField.Field.Interface()))
		}
		primaryValues = append(primaryValues, fmt.Sprint(rand.Intn(1000)))
		return utils.ToParamString(url.QueryEscape(strings.Join(primaryValues, "_")))
	}
	return fmt.Sprint(value)
}

func (context *Context) isNewRecord(value interface{}) bool {
	if value == nil {
		return true
	}
	return context.DB.NewRecord(value)
}

func (context *Context) newResourcePath(res *Resource) string {
	return path.Join(context.URLFor(res), "new")
}

// RoutePrefix return route prefix of resource
func (res *Resource) RoutePrefix() string {
	var params string
	for res.ParentResource != nil {
		params = path.Join(res.ParentResource.ToParam(), res.ParentResource.ParamIDName(), params)
		res = res.ParentResource
	}
	return params
}

// UseTheme append used themes into current context, will load those theme's stylesheet, javascripts in admin pages
func (context *Context) UseTheme(name string) {
	context.usedThemes = append(context.usedThemes, name)
}

// URLFor generate url for resource value
//     context.URLFor(&Product{})
//     context.URLFor(&Product{ID: 111})
//     context.URLFor(productResource)
func (context *Context) URLFor(value interface{}, resources ...*Resource) string {
	getPrefix := func(res *Resource) string {
		var params string
		for res.ParentResource != nil {
			params = path.Join(res.ParentResource.ToParam(), res.ParentResource.GetPrimaryValue(context.Request), params)
			res = res.ParentResource
		}
		return params
	}

	if admin, ok := value.(*Admin); ok {
		return context.GenURL(admin.router.Prefix)
	} else if urler, ok := value.(interface {
		ToURLString(*Context) string
	}); ok {
		return urler.ToURLString(context)
	} else if res, ok := value.(*Resource); ok {
		return context.GenURL(getPrefix(res), res.ToParam())
	} else {
		var res *Resource

		if len(resources) > 0 {
			res = resources[0]
		}

		if res == nil {
			res = context.Admin.GetResource(reflect.Indirect(reflect.ValueOf(value)).Type().String())
		}

		if res != nil {
			if res.Config.Singleton {
				return context.GenURL(getPrefix(res), res.ToParam())
			}

			var (
				scope         = context.DB.NewScope(value)
				primaryFields []string
				primaryValues = map[string]string{}
			)

			for _, primaryField := range res.PrimaryFields {
				if field, ok := scope.FieldByName(primaryField.Name); ok {
					primaryFields = append(primaryFields, fmt.Sprint(field.Field.Interface())) // TODO improve me
				}
			}

			for _, field := range scope.PrimaryFields() {
				useAsPrimaryField := false
				for _, primaryField := range res.PrimaryFields {
					if field.DBName == primaryField.DBName {
						useAsPrimaryField = true
						break
					}
				}

				if !useAsPrimaryField {
					primaryValues[fmt.Sprintf("primary_key[%v_%v]", scope.TableName(), field.DBName)] = fmt.Sprint(reflect.Indirect(field.Field).Interface())
				}
			}

			urlPath := path.Join(res.ToParam(), strings.Join(primaryFields, ","))

			if len(primaryValues) > 0 {
				var primaryValueParams []string
				for key, value := range primaryValues {
					primaryValueParams = append(primaryValueParams, fmt.Sprintf("%v=%v", key, url.QueryEscape(value)))
				}
				urlPath = urlPath + "?" + strings.Join(primaryValueParams, "&")
			}
			return context.GenURL(getPrefix(res), urlPath)
		}
	}
	return context.GenURL("")
}

func (context *Context) linkTo(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		if linkStr[0:1] == "@" {
			linkStr = context.GenURL(linkStr[1:])
		}
		return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, context.URLFor(link), text))
}

func (context *Context) valueOf(valuer func(interface{}, *qor.Context) interface{}, value interface{}, meta *Meta) interface{} {
	if valuer != nil {
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() != reflect.Ptr {
			reflectPtr := reflect.New(reflectValue.Type())
			reflectPtr.Elem().Set(reflectValue)
			value = reflectPtr.Interface()
		}

		result := valuer(value, context.Context)

		if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
			if reflectValue.Kind() == reflect.Ptr {
				if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
					return nil
				}

				result = reflectValue.Elem().Interface()
			}

			if meta.Type == "number" || meta.Type == "float" {
				if context.isNewRecord(value) && equal(reflect.Zero(reflect.TypeOf(result)).Interface(), result) {
					return nil
				}
			}
			return result
		}
		return nil
	}

	utils.ExitWithMsg(fmt.Sprintf("No valuer found for meta %v of resource %v", meta.Name, meta.baseResource.Name))
	return nil
}

// RawValueOf return raw value of a meta for current resource
func (context *Context) RawValueOf(value interface{}, meta *Meta) interface{} {
	return context.valueOf(meta.GetValuer(), value, meta)
}

// FormattedValueOf return formatted value of a meta for current resource
func (context *Context) FormattedValueOf(value interface{}, meta *Meta) interface{} {
	result := context.valueOf(meta.GetFormattedValuer(), value, meta)
	if resultValuer, ok := result.(driver.Valuer); ok {
		if result, err := resultValuer.Value(); err == nil {
			return result
		}
	}

	return result
}

func (context *Context) renderForm(value interface{}, sections []*Section) template.HTML {
	var result = bytes.NewBufferString("")
	context.renderSections(value, sections, []string{"QorResource"}, result, "form")
	return template.HTML(result.String())
}

func (context *Context) renderSections(value interface{}, sections []*Section, prefix []string, writer *bytes.Buffer, kind string) {
	for _, section := range sections {
		var rows []struct {
			Length      int
			ColumnsHTML template.HTML
		}

		for _, column := range section.Rows {
			columnsHTML := bytes.NewBufferString("")
			var exclude int
			for _, col := range column {
				meta := section.Resource.GetMeta(col)
				if meta != nil {
					if meta.Enabled == nil || meta.Enabled(context, meta) {
						context.renderMeta(meta, value, prefix, kind, columnsHTML)
					} else {
						exclude++
					}
				}
			}

			rows = append(rows, struct {
				Length      int
				ColumnsHTML template.HTML
			}{
				Length:      len(column) - exclude,
				ColumnsHTML: template.HTML(string(columnsHTML.Bytes())),
			})
		}

		var data = map[string]interface{}{
			"Section": section,
			"Title":   template.HTML(section.Title),
			"Rows":    rows,
		}

		if executor, err := context.GetTemplate("metas/section"); err == nil {
			executor.Execute(writer, data, context.FuncValues())
		}
	}
}

func (context *Context) renderFilter(filter *Filter) template.HTML {
	var (
		err    error
		result = bytes.NewBufferString("")
	)

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			result.WriteString(fmt.Sprintf("Get error when render template for filter %v (%v): %v", filter.Name, filter.Type, r))
		}
	}()

	if executor, err := context.GetTemplate(fmt.Sprintf("metas/filter/%v", filter.Type)); err == nil {
		var data = map[string]interface{}{
			"Filter":          filter,
			"Label":           filter.Label,
			"InputNamePrefix": fmt.Sprintf("filters[%v]", filter.Name),
			"Context":         context,
			"Resource":        context.Resource,
		}

		err = executor.Execute(result, data, context.FuncValues())
	}

	if err != nil {
		result.WriteString(fmt.Sprintf("got error when render filter template for %v(%v):%v", filter.Name, filter.Type, err))
	}

	return template.HTML(result.String())
}

func (context *Context) renderMeta(meta *Meta, value interface{}, prefix []string, metaType string, writer *bytes.Buffer) {
	var (
		err      error
		funcsMap = funcs.FuncMap{}
		executor *template.Executor
	)
	prefix = append(prefix, meta.Name)

	var generateNestedRenderSections = func(kind string) func(interface{}, []*Section, int) template.HTML {
		return func(value interface{}, sections []*Section, index int) template.HTML {
			var result = bytes.NewBufferString("")
			var newPrefix = append([]string{}, prefix...)

			if index >= 0 {
				last := newPrefix[len(newPrefix)-1]
				newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index))
			}

			if len(sections) > 0 {
				for _, field := range context.DB.NewScope(value).PrimaryFields() {
					if meta := sections[0].Resource.GetMeta(field.Name); meta != nil {
						context.renderMeta(meta, value, newPrefix, kind, result)
					}
				}

				context.renderSections(value, sections, newPrefix, result, kind)
			}

			return template.HTML(result.String())
		}
	}

	funcsMap["render_nested_form"] = generateNestedRenderSections("form")

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			msg := fmt.Sprintf("Get error when render template for meta %v (%v): %v", meta.Name, meta.Type, r)
			writer.WriteString(msg)
			println(msg)
		}
	}()

	switch {
	case meta.Config != nil:
		if templater, ok := meta.Config.(interface {
			GetTemplate(context *Context, metaType string) (*template.Executor, error)
		}); ok {
			if executor, err = templater.GetTemplate(context, metaType); err == nil {
				break
			}
		}
		fallthrough
	default:
		var others []string
		metaUserType := meta.GetType(context)

		if metaUserType != "" {
			others = append(others, fmt.Sprintf("metas/%v/%v", metaType, metaUserType))
		}

		if executor, err = context.GetTemplateOrDefault(fmt.Sprintf("%v/metas/%v/%v", meta.baseResource.ToParam(), metaType, meta.Name),
			TemplateExecutorMetaValue, others...); err != nil {
			err = fmt.Errorf("haven't found %v template for meta %v: %v", metaType, meta.Name, err)
		}
	}

	if err == nil {
		var scope = context.DB.NewScope(value)
		var data = map[string]interface{}{
			"Context":       context,
			"BaseResource":  meta.baseResource,
			"Meta":          meta,
			"ResourceValue": value,
			"Value":         context.FormattedValueOf(value, meta),
			"Label":         meta.Label,
			"InputName":     strings.Join(prefix, "."),
		}

		if !scope.PrimaryKeyZero() {
			data["InputId"] = utils.ToParamString(fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), meta.Name))
		}

		data["CollectionValue"] = func() [][]string {
			fmt.Printf("%v: Call .CollectionValue from views already Deprecated, get the value with `.Meta.Config.GetCollection .ResourceValue .Context`", meta.Name)
			return meta.Config.(interface {
				GetCollection(value interface{}, context *Context) [][]string
			}).GetCollection(value, context)
		}

		err = executor.Execute(writer, data, context.FuncValues(), funcsMap)
	}

	if err != nil {
		msg := fmt.Sprintf("got error when render %v template for %v(%v): %v", metaType, meta.Name, meta.Type, err)
		fmt.Fprint(writer, msg)
		utils.ExitWithMsg(msg)
	}
}

func (context *Context) isEqual(value interface{}, hasValue interface{}) bool {
	var result string

	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: hasValue}
		result = fmt.Sprint(scope.PrimaryKeyValue())
	} else {
		result = fmt.Sprint(hasValue)
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		return fmt.Sprint(scope.PrimaryKeyValue()) == result
	} else if reflectValue.Kind() == reflect.String {
		return reflectValue.Interface().(string) == result
	} else {
		return fmt.Sprint(reflectValue.Interface()) == result
	}
}

func (context *Context) isIncluded(value interface{}, hasValue interface{}) bool {
	var result string
	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: hasValue}
		result = fmt.Sprint(scope.PrimaryKeyValue())
	} else {
		result = fmt.Sprint(hasValue)
	}

	primaryKeys := []interface{}{}
	reflectValue := reflect.Indirect(reflect.ValueOf(value))

	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			if value := reflectValue.Index(i); value.IsValid() {
				if reflect.Indirect(value).Kind() == reflect.Struct {
					scope := &gorm.Scope{Value: reflectValue.Index(i).Interface()}
					primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
				} else {
					primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue.Index(i)).Interface())
				}
			}
		}
	} else if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
	} else if reflectValue.Kind() == reflect.String {
		return strings.Contains(reflectValue.Interface().(string), result)
	} else if reflectValue.IsValid() {
		primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue).Interface())
	}

	for _, key := range primaryKeys {
		if fmt.Sprint(key) == result {
			return true
		}
	}
	return false
}

func (context *Context) getResource(resources ...*Resource) *Resource {
	for _, res := range resources {
		return res
	}
	return context.Resource
}

func (context *Context) indexSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.IndexAttrs(), context, roles.Read)
}

func (context *Context) editSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.EditAttrs(), context, roles.Read)
}

func (context *Context) newSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.NewAttrs(), context, roles.Create)
}

func (context *Context) showSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.ShowAttrs(), context, roles.Read)
}

type menu struct {
	*Menu
	Active   bool
	SubMenus []*menu
}

func (context *Context) getMenus() (menus []*menu) {
	var (
		globalMenu        = &menu{}
		mostMatchedMenu   *menu
		mostMatchedLength int
		addMenu           func(*menu, []*Menu)
		path              = context.Path()
	)

	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			if m.HasPermission(roles.Read, context.Context) {
				var menu = &menu{Menu: m}
				url := m.URL(context)

				if url[0:1] == "@" {
					url = url[1:]
				}

				if strings.HasPrefix(path, url) && len(url) > mostMatchedLength {
					mostMatchedMenu = menu
					mostMatchedLength = len(url)
				}

				addMenu(menu, menu.GetSubMenus())
				parent.SubMenus = append(parent.SubMenus, menu)
			}
		}
	}

	addMenu(globalMenu, context.Admin.GetMenus())

	if context.Action != "search_center" && mostMatchedMenu != nil {
		mostMatchedMenu.Active = true
	}

	return globalMenu.SubMenus
}

type scope struct {
	*Scope
	Active bool
}

type scopeMenu struct {
	Group  string
	Scopes []scope
}

// GetScopes get scopes from current context
func (context *Context) GetScopes() (menus []*scopeMenu) {
	if context.Resource == nil {
		return
	}

	scopes := context.Request.URL.Query()["scopes"]
OUT:
	for _, s := range context.Resource.scopes {
		if s.Visible != nil && !s.Visible(context) {
			continue
		}

		menu := scope{Scope: s}

		for _, s := range scopes {
			if s == menu.Name {
				menu.Active = true
			}
		}

		if !menu.Default {
			if menu.Group != "" {
				for _, m := range menus {
					if m.Group == menu.Group {
						m.Scopes = append(m.Scopes, menu)
						continue OUT
					}
				}
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			} else {
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			}
		}
	}
	return menus
}

// HasPermissioner has permission interface
type HasPermissioner interface {
	HasPermission(roles.PermissionMode, *qor.Context) bool
}

func (context *Context) hasCreatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Create, context.Context)
}

func (context *Context) hasReadPermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Read, context.Context)
}

func (context *Context) hasUpdatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Update, context.Context)
}

func (context *Context) hasDeletePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Delete, context.Context)
}

// Page contain pagination information
type Page struct {
	Page       int
	Current    bool
	IsPrevious bool
	IsNext     bool
	IsFirst    bool
	IsLast     bool
}

type PaginationResult struct {
	Pagination Pagination
	Pages      []Page
}

const visiblePageCount = 8

// Pagination return pagination information
// Keep visiblePageCount's pages visible, exclude prev and next link
// Assume there are 12 pages in total.
// When current page is 1
// [current, 2, 3, 4, 5, 6, 7, 8, next]
// When current page is 6
// [prev, 2, 3, 4, 5, current, 7, 8, 9, 10, next]
// When current page is 10
// [prev, 5, 6, 7, 8, 9, current, 11, 12]
// If total page count less than VISIBLE_PAGE_COUNT, always show all pages
func (context *Context) Pagination() *PaginationResult {
	var (
		pages      []Page
		pagination = context.Searcher.Pagination
		pageCount  = pagination.PerPage
	)

	if pageCount == 0 {
		if context.Resource != nil && context.Resource.Config.PageCount != 0 {
			pageCount = context.Resource.Config.PageCount
		} else {
			pageCount = PaginationPageCount
		}
	}

	if pagination.Total <= pageCount && pagination.CurrentPage <= 1 {
		return nil
	}

	start := pagination.CurrentPage - visiblePageCount/2
	if start < 1 {
		start = 1
	}

	end := start + visiblePageCount - 1 // -1 for "start page" itself
	if end > pagination.Pages {
		end = pagination.Pages
	}

	if (end-start) < visiblePageCount && start != 1 {
		start = end - visiblePageCount + 1
	}
	if start < 1 {
		start = 1
	}

	// Append prev link
	if start > 1 {
		pages = append(pages, Page{Page: 1, IsFirst: true})
		pages = append(pages, Page{Page: pagination.CurrentPage - 1, IsPrevious: true})
	}

	for i := start; i <= end; i++ {
		pages = append(pages, Page{Page: i, Current: pagination.CurrentPage == i})
	}

	// Append next link
	if end < pagination.Pages {
		pages = append(pages, Page{Page: pagination.CurrentPage + 1, IsNext: true})
		pages = append(pages, Page{Page: pagination.Pages, IsLast: true})
	}

	return &PaginationResult{Pagination: pagination, Pages: pages}
}

func (context *Context) themesClass() (result string) {
	var results = map[string]bool{}
	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			if strings.HasPrefix(theme.GetName(), "-") {
				results[strings.TrimPrefix(theme.GetName(), "-")] = false
			} else if _, ok := results[theme.GetName()]; !ok {
				results[theme.GetName()] = true
			}
		}
	}

	var names []string
	for name, enabled := range results {
		if enabled {
			names = append(names, "qor-theme-"+name)
		}
	}
	return strings.Join(names, " ")
}

func (context *Context) javaScriptTag(names ...string) template.HTML {
	var results []string
	prefix := context.GenStaticURL(path.Join("assets", "javascripts"))
	for _, name := range names {
		results = append(results, fmt.Sprintf(`<script src="%s/%s.js"></script>`, prefix, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := context.GenStaticURL(path.Join("assets", "stylesheets"))
	for _, name := range names {
		results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s/%s.css">`, prefix, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) getThemeNames() (themes []string) {
	themesMap := map[string]bool{}

	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			if _, ok := themesMap[theme.GetName()]; !ok {
				themes = append(themes, theme.GetName())
			}
		}
	}

	for _, usedTheme := range context.usedThemes {
		if _, ok := themesMap[usedTheme]; !ok {
			themes = append(themes, usedTheme)
		}
	}

	return
}

func (context *Context) loadThemeStyleSheets() template.HTML {
	var results []string
	for _, themeName := range context.getThemeNames() {
		var file = path.Join("themes", themeName, "assets", "stylesheets", themeName+".css")
		if _, err := context.Asset(file); err == nil {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s?theme=%s">`, context.GenStaticURL(file), themeName))
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadThemeJavaScripts() template.HTML {
	var results []string
	for _, themeName := range context.getThemeNames() {
		var file = path.Join("themes", themeName, "assets", "javascripts", themeName+".js")
		if _, err := context.Asset(file); err == nil {
			results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, context.GenStaticURL(file), themeName))
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadAdminJavaScripts() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "javascripts", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".js")
	if _, err := context.Asset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, context.GenStaticURL(file)))
	}
	return ""
}

func (context *Context) loadAdminStyleSheets() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "stylesheets", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".css")

	if _, err := context.Asset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, context.GenStaticURL(file)))
	}
	return ""
}

func (context *Context) loadActions(action string) template.HTML {
	var (
		actionPatterns, actionKeys, actionFiles []string
		actions                                 = map[string]string{}
	)

	switch action {
	case "index", "show", "edit", "new":
		actionPatterns = []string{filepath.Join("actions", action, "*.tmpl"), "actions/*.tmpl"}

		if !context.Resource.isSetShowAttrs && action == "edit" {
			actionPatterns = []string{filepath.Join("actions", "show", "*.tmpl"), "actions/*.tmpl"}
		}
	case "global":
		actionPatterns = []string{"actions/*.tmpl"}
	default:
		actionPatterns = []string{filepath.Join("actions", action, "*.tmpl")}
	}

	for _, pattern := range actionPatterns {
		for _, themeName := range context.getThemeNames() {
			if resourcePath := context.resourcePath(); resourcePath != "" {
				if matches, err := context.Admin.AssetFS.Glob(filepath.Join("themes", themeName, resourcePath, pattern)); err == nil {
					actionFiles = append(actionFiles, matches...)
				}
			}

			if matches, err := context.Admin.AssetFS.Glob(filepath.Join("themes", themeName, pattern)); err == nil {
				actionFiles = append(actionFiles, matches...)
			}
		}

		if resourcePath := context.resourcePath(); resourcePath != "" {
			if matches, err := context.Admin.AssetFS.Glob(filepath.Join(resourcePath, pattern)); err == nil {
				actionFiles = append(actionFiles, matches...)
			}
		}

		if matches, err := context.Admin.AssetFS.Glob(pattern); err == nil {
			actionFiles = append(actionFiles, matches...)
		}
	}

	// before files have higher priority
	for _, actionFile := range actionFiles {
		actionFile = strings.TrimSuffix(actionFile, ".tmpl")
		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(path.Base(actionFile), "")

		if _, ok := actions[base]; !ok {
			actionKeys = append(actionKeys, path.Base(actionFile))
			actions[base] = actionFile
		}
	}

	sort.Strings(actionKeys)

	var (
		result   = bytes.NewBufferString("")
		err      error
		executor *template.Executor
	)

	for _, key := range actionKeys {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Sprintf("Get error when render action %v: %v", key, r)
				utils.ExitWithMsg(err)
				result.WriteString(err)
			}
		}()

		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(key, "")
		if executor, err = context.GetTemplate(actions[base]); err == nil {
			err = executor.Execute(result, context, context.FuncValues())
		}
		if err != nil {
			result.WriteString(err.Error())
			utils.ExitWithMsg(err)
			return template.HTML("")
		}
	}

	return template.HTML(strings.TrimSpace(result.String()))
}

func (context *Context) logoutURL() string {
	if context.Admin.Auth != nil {
		return context.Admin.Auth.LogoutURL(context)
	}
	return ""
}

func (context *Context) t(key string, defaul ... interface{}) template.HTML {
	var defauls []string
	for _, value := range defaul {
		defauls = append(defauls, fmt.Sprint(value))
	}
	return context.T(key, defauls...)
}

func (context *Context) tt(key string, data interface{}, defaul ... interface{}) template.HTML {
	var defauls []string
	for _, value := range defaul {
		defauls = append(defauls, fmt.Sprint(value))
	}
	return context.TT(key, data, defauls...)
}

func (context *Context) isSortableMeta(meta *Meta) bool {
	for _, attr := range context.Resource.SortableAttrs() {
		if attr == meta.Name && meta.FieldStruct != nil && meta.FieldStruct.IsNormal && meta.FieldStruct.DBName != "" {
			return true
		}
	}
	return false
}

func (context *Context) convertSectionToMetas(res *Resource, sections []*Section) []*Meta {
	return res.ConvertSectionToMetas(sections)
}

type formatedError struct {
	Label  string
	Errors []string
}

func (context *Context) getFormattedErrors() (formatedErrors []formatedError) {
	type labelInterface interface {
		Label() string
	}

	for _, err := range context.GetErrors() {
		if labelErr, ok := err.(labelInterface); ok {
			var found bool
			label := labelErr.Label()
			for _, formatedError := range formatedErrors {
				if formatedError.Label == label {
					formatedError.Errors = append(formatedError.Errors, err.Error())
				}
			}
			if !found {
				formatedErrors = append(formatedErrors, formatedError{Label: label, Errors: []string{err.Error()}})
			}
		} else {
			formatedErrors = append(formatedErrors, formatedError{Errors: []string{err.Error()}})
		}
	}
	return
}

// AllowedActions return allowed actions based on context
func (context *Context) AllowedActions(actions []*Action, mode string, records ...interface{}) []*Action {
	var allowedActions []*Action
	for _, action := range actions {
		for _, m := range action.Modes {
			if m == mode {
				var permission = roles.Update
				switch strings.ToUpper(action.Method) {
				case "POST":
					permission = roles.Create
				case "DELETE":
					permission = roles.Delete
				case "PUT":
					permission = roles.Update
				case "GET":
					permission = roles.Read
				}

				if action.IsAllowed(permission, context, records...) {
					allowedActions = append(allowedActions, action)
					break
				}
			}
		}
	}
	return allowedActions
}

func (context *Context) pageTitle() template.HTML {
	if context.Action == "search_center" {
		return context.t("qor_admin.search_center.title", "Search Center")
	}

	if context.Resource == nil {
		if context.PageTitle != "" {
			return context.t(context.PageTitle)
		}
		return context.t("qor_admin.layout.title", "Admin")
	}

	if context.Action == "action" {
		if action, ok := context.Result.(*Action); ok {
			return context.Resource.GetActionLabel(context, action)
		}
	}

	var (
		defaultValue string
		titleKey     = fmt.Sprintf("qor_admin.form.%v.title", context.Action)
		usePlural    bool
	)

	defaultValue = context.GetActionLabel()

	if defaultValue == "" {
		defaultValue = "{{.}}"
		if !context.Resource.Config.Singleton {
			usePlural = true
		}
	}

	resourceName := context.Resource.GetLabel(context, usePlural)
	title := fmt.Sprint(context.t(titleKey, defaultValue))

	return utils.RenderHtmlTemplate(title, resourceName)
}

// FuncValues return funcs FuncValues
func (context *Context) FuncValues() *funcs.FuncValues {
	if context.funcValues == nil {
		v, err := funcs.CreateValuesFunc(context.FuncMaps()...)
		if err != nil {
			panic(err)
		}
		context.funcValues = v
	}
	return context.funcValues
}

// FuncMap return funcs map
func (context *Context) FuncMaps() []funcs.FuncMap {
	funcMaps := []template.FuncMap{
		template.FuncMap{
			"qor_context":          func() *qor.Context { return context.Context },
			"site":                 func() qor.SiteInterface { return context.Context.Site },
			"public_url":           func(args ...string) string { return context.Context.Site.PublicURL() },
			"public_urlf":          func(args ...interface{}) string { return context.Context.Site.PublicURLf(args...) },
			"admin_context":        func() *Context { return context },
			"current_user":         func() qor.CurrentUser { return context.CurrentUser },
			"get_resource":         context.Admin.GetResource,
			"new_resource_context": context.NewResourceContext,
			"is_new_record":        context.isNewRecord,
			"is_equal":             context.isEqual,
			"is_included":          context.isIncluded,
			"primary_key_of":       context.primaryKeyOf,
			"unique_key_of":        context.uniqueKeyOf,
			"formatted_value_of":   context.FormattedValueOf,
			"raw_value_of":         context.RawValueOf,

			"t":          context.t,
			"tt":         context.tt,
			"flashes":    func() []session.Message { return context.SessionManager().Flashes() },
			"pagination": context.Pagination,
			"escape":     html.EscapeString,
			"raw":        func(str string) template.HTML { return template.HTML(utils.HTMLSanitizer.Sanitize(str)) },
			"unsafe_raw": func(str string) template.HTML { return template.HTML(str) },
			"equal":      equal,
			"stringify":  utils.Stringify,
			"lower": func(value interface{}) string {
				return strings.ToLower(fmt.Sprint(value))
			},
			"plural": func(value interface{}) string {
				return inflection.Plural(fmt.Sprint(value))
			},
			"singular": func(value interface{}) string {
				return inflection.Singular(fmt.Sprint(value))
			},
			"marshal": func(v interface{}) template.JS {
				switch value := v.(type) {
				case string:
					return template.JS(value)
				case template.HTML:
					return template.JS(value)
				default:
					byt, _ := json.Marshal(v)
					return template.JS(byt)
				}
			},

			"render":      context.Render,
			"render_text": context.renderText,
			"render_with": context.renderWith,
			"render_form": context.renderForm,
			"render_meta": func(value interface{}, meta *Meta, types ...string) template.HTML {
				var (
					result = bytes.NewBufferString("")
					typ    = "index"
				)

				for _, t := range types {
					typ = t
				}

				context.renderMeta(meta, value, []string{}, typ, result)
				return template.HTML(result.String())
			},
			"render_filter": context.renderFilter,
			"page_title":    context.pageTitle,
			"meta_label": func(meta *Meta) template.HTML {
				key, defaul := meta.GetLabelPair()
				return context.Admin.T(context.Context, key, defaul)
			},
			"meta_placeholder": func(meta *Meta, context *Context, placeholder string) template.HTML {
				if getPlaceholder, ok := meta.Config.(interface {
					GetPlaceholder(*Context) (template.HTML, bool)
				}); ok {
					if str, ok := getPlaceholder.GetPlaceholder(context); ok {
						return str
					}
				}

				key := fmt.Sprintf("%v.attributes.%v.placeholder", meta.baseResource.I18nPrefix, meta.Name)
				return context.Admin.T(context.Context, key, placeholder)
			},

			"url_for":            context.URLFor,
			"link_to":            context.linkTo,
			"patch_current_url":  context.PatchCurrentURL,
			"patch_url":          context.PatchURL,
			"join_current_url":   context.JoinCurrentURL,
			"join_url":           context.JoinURL,
			"logout_url":         context.logoutURL,
			"search_center_path": func() string { return context.JoinPath("!search") },
			"new_resource_path":  context.newResourcePath,
			"defined_resource_show_page": func(res *Resource) bool {
				if res != nil {
					if r := context.Admin.GetResource(res.Name); r != nil {
						return r.isSetShowAttrs
					}
				}

				return false
			},
			"get_menus":                 context.getMenus,
			"get_scopes":                context.GetScopes,
			"get_formatted_errors":      context.getFormattedErrors,
			"load_actions":              context.loadActions,
			"allowed_actions":           context.AllowedActions,
			"is_sortable_meta":          context.isSortableMeta,
			"index_sections":            context.indexSections,
			"show_sections":             context.showSections,
			"new_sections":              context.newSections,
			"edit_sections":             context.editSections,
			"convert_sections_to_metas": context.convertSectionToMetas,

			"has_create_permission": context.hasCreatePermission,
			"has_read_permission":   context.hasReadPermission,
			"has_update_permission": context.hasUpdatePermission,
			"has_delete_permission": context.hasDeletePermission,

			"qor_theme_class":        context.themesClass,
			"javascript_tag":         context.javaScriptTag,
			"stylesheet_tag":         context.styleSheetTag,
			"load_theme_stylesheets": context.loadThemeStyleSheets,
			"load_theme_javascripts": context.loadThemeJavaScripts,
			"load_admin_stylesheets": context.loadAdminStyleSheets,
			"load_admin_javascripts": context.loadAdminJavaScripts,

			"global_url":       context.GenGlobalURL,
			"static_url":       context.GenGlobalStaticURL,
			"url":              context.GenURL,
			"admin_static_url": context.GenStaticURL,
			"locale": func() string {
				return context.Locale
			},
		},
		context.Admin.funcMaps,
	}

	funcMaps = append(funcMaps, context.funcMaps...)

	return funcMaps
}
