package admin

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/ecletus/auth"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/tracederror"
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/go-aorm/aorm"
)

var TemplateExecutorMetaValue = template.Must(template.New(PKG + ".meta_value").Parse(`{{if and .MetaValue .MetaValue.Severity -}}
    <span class="severity_{{.MetaValue.Severity}} severity--text severity--bg">{{.Value}}</span>
{{- else -}}
    {{.Value}}
{{- end}}`)).CreateExecutor()

func (this *Context) primaryKeyOf(value interface{}) interface{} {
	if idGetter, ok := value.(interface{ GetID() aorm.ID }); ok {
		return idGetter.GetID()
	}
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		return aorm.IdOf(value)
	}
	return fmt.Sprint(value)
}

func (this *Context) uniqueKeyOf(value interface{}) interface{} {
	if reflect.Indirect(reflect.ValueOf(value)).Kind() == reflect.Struct {
		var primaryValues []string
		for _, primaryField := range aorm.PrimaryFieldsOf(value) {
			primaryValues = append(primaryValues, fmt.Sprint(primaryField.Field.Interface()))
		}
		primaryValues = append(primaryValues, fmt.Sprint(rand.Intn(1000)))
		return utils.ToParamString(url.QueryEscape(strings.Join(primaryValues, "_")))
	}
	return fmt.Sprint(value)
}

func (this *Context) isNewRecord(value interface{}) bool {
	if value == nil {
		return true
	} else if indirectType(reflect.TypeOf(value)).Kind() != reflect.Struct {
		return false
	}
	struc := aorm.StructOf(value)
	if struc == nil || struc.Parent != nil {
		return false
	}
	id := struc.GetID(value)
	return id == nil || id.IsZero()
}

func (this *Context) SetNewResourceForPath(res *Resource) {
	this.SetValue(PKG+".new_resource_path", res)
}

func (this *Context) newResourcePath(res *Resource) string {
	if res2 := this.Value(PKG + ".new_resource_path"); res2 != nil {
		res = res2.(*Resource)
	}
	return res.GetContextIndexURI(this) + "/new"
}

func (this *Context) linkTo(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		if linkStr == "" {
			linkStr = "javascript:void(0)"
		} else if linkStr[0:1] == "@" {
			linkStr = this.Path(linkStr[1:])
		}
		return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, this.URLFor(link), text))
}

func (this *Context) linkToAjaxLoad(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		if linkStr == "" {
			linkStr = "javascript:void(0)"
		} else if linkStr[0:1] == "@" {
			linkStr = this.Path(linkStr[1:])
		}
		return template.HTML(fmt.Sprintf(`<a href="%v" data-url="%v">%v</a>`, linkStr, linkStr, text))
	}
	url := this.URLFor(link)
	return template.HTML(fmt.Sprintf(`<a href="%v" data-url="%v">%v</a>`, url, url, text))
}

func (this *Context) valueOf(valuer func(interface{}, *core.Context) interface{}, value interface{}, meta *Meta) interface{} {
	if valuer != nil {
		if value == nil {
			return nil
		}

		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() != reflect.Ptr {
			reflectPtr := reflect.New(reflectValue.Type())
			reflectPtr.Elem().Set(reflectValue)
			value = reflectPtr.Interface()
		}

		originalResult := valuer(value, this.Context)
		result := originalResult
		if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
			if reflectValue.Kind() == reflect.Ptr {
				if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
					return nil
				}
				reflectValue = reflectValue.Elem()
				result = reflectValue.Interface()
			}

			if meta.Type == "number" || meta.Type == "float" {
				if this.isNewRecord(value) && equal(reflect.Zero(reflectValue.Type()).Interface(), result) {
					return nil
				}
			} else if ID, ok := result.(aorm.ID); ok {
				return ID.String()
			}
			return originalResult
		}
		return nil
	}

	if meta.Virtual {
		panic(fmt.Errorf("No valuer found for meta %v of resource %v", meta.Name, meta.BaseResource.Name))
	}
	return nil
}

func (this *Context) isEqual(value interface{}, hasValue interface{}) bool {
	if (value == nil && hasValue != nil) || (value != nil && hasValue == nil) {
		return false
	}

	if equaler, ok := value.(Equaler); ok {
		return equaler.Equals(hasValue)
	}

	var (
		result          string
		reflectHasValue = reflect.Indirect(reflect.ValueOf(hasValue))
	)

	if reflectHasValue.Kind() == reflect.Struct {
		result = aorm.IdOf(hasValue).String()
	} else {
		result = fmt.Sprint(hasValue)
	}

	switch vt := value.(type) {
	case Equaler:
		return vt.Equals(result)
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.Kind() == reflect.Struct {
		return aorm.IdOf(hasValue).String() == result
	}

	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	if reflectHasValue.Type().ConvertibleTo(reflectValue.Type()) {
		if reflect.DeepEqual(reflectValue.Interface(), reflectHasValue.Convert(reflectValue.Type()).Interface()) {
			return true
		}
	}

	return fmt.Sprint(reflectValue.Interface()) == result
}

func (this *Context) isIncluded(value interface{}, hasValue interface{}) bool {
	var (
		result       string
		primaryKeys  []interface{}
		reflectValue = reflect.Indirect(reflect.ValueOf(value))
	)
	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		result = aorm.IdOf(hasValue).String()
	} else {
		result = fmt.Sprint(hasValue)
	}

	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			if value := reflectValue.Index(i); value.IsValid() {
				var item interface{}
				if reflect.Indirect(value).Kind() == reflect.Struct {
					item = reflectValue.Index(i).Interface()
				} else {
					item = reflect.Indirect(reflectValue.Index(i)).Interface()
				}
				primaryKeys = append(primaryKeys, this.primaryKeyOf(item))
			}
		}
	} else if reflectValue.Kind() == reflect.Struct {
		primaryKeys = append(primaryKeys, aorm.IdOf(value))
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

func (this *Context) getResource(resources ...*Resource) *Resource {
	for _, res := range resources {
		return res
	}
	return this.Resource
}

func (this *Context) indexSections() (sections Sections) {
	if this.Layout != "" && this.Layout != "index" {
		layout := this.Resource.GetAdminLayout(this.Layout)
		filter := new(SectionsFilter)

		if layout.NotIndexRenderID && !this.Api {
			filter.SetExcludes(layout.MetaID)
		}

		sections = layout.GetSections(this.Resource, this, nil, filter)
	} else {
		scheme := this.Scheme
		if scheme == nil {
			scheme = this.Resource.Scheme
		}
		sections = scheme.IndexSections(this)
	}

	sections = sections.Allowed(nil, this, roles.Read)
	return sections
}

func (this *Context) editSections(res *Resource, recorde ...interface{}) []*Section {
	if len(recorde) == 0 {
		recorde = append(recorde, nil)
	}
	return res.EditSections(this, recorde[0])
}

func (this *Context) newSections(res *Resource) []*Section {
	secs := res.NewSections(this)
	return secs
}

func (this *Context) showSections() []*Section {
	return this.Resource.ShowSections(this, this.ResourceRecord)
}

func (this *Context) editMetaSections(meta *Meta) []*Section {
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.EditSections(this, this.ResourceRecord)
		}
		if sc, ok := meta.Config.(*CollectionEditConfig); ok && sc.Sections != nil {
			attrs = sc.EditSections(this, this.ResourceRecord)
		}
	}
	if len(attrs) == 0 {
		attrs = res.EditAttrs()
	}
	secs := Sections(attrs).Allowed(this.ResourceRecord, this, roles.Update)
	return secs
}

func (this *Context) newMetaSections(meta *Meta) []*Section {
	if meta.Resource == nil {
		panic("admin.func_map.newMetaSections: meta.Resource is nil")
	}
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.NewSections(this)
		}
		if sc, ok := meta.Config.(*CollectionEditConfig); ok && sc.Sections != nil {
			attrs = sc.NewSections(this)
		}
	}
	if len(attrs) == 0 {
		attrs = res.NewAttrs()
	}
	return Sections(attrs).Allowed(this.ResourceRecord, this, roles.Create)
}

func (this *Context) showMetaSections(meta *Meta) []*Section {
	if meta.Resource == nil {
		panic("admin.func_map.showMetaSections: meta.Resource is nil")
	}
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.ShowSections(this)
		}
		if sc, ok := meta.Config.(*CollectionEditConfig); ok && sc.Sections != nil {
			attrs = sc.ShowSections(this, this.ResourceRecord)
		}
	}
	if len(attrs) == 0 {
		attrs = res.ShowSections(this, this.ResourceRecord)
	}
	return Sections(attrs).Allowed(this.ResourceRecord, this, roles.Read)
}

func (this *Context) themesClass() (result string) {
	var results = map[string]bool{}
	if this.Resource != nil {
		for _, theme := range this.Resource.Config.Themes {
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

func (this *Context) getThemeNames() (themes []string) {
	themesMap := map[string]bool{}

	if this.Resource != nil {
		for _, theme := range this.Resource.Config.Themes {
			if _, ok := themesMap[theme.GetName()]; !ok {
				themes = append(themes, theme.GetName())
			}
		}
	}

	for _, usedTheme := range this.usedThemes {
		if _, ok := themesMap[usedTheme]; !ok {
			themes = append(themes, usedTheme)
		}
	}

	return
}

func (this *Context) loadThemeStyleSheets() template.HTML {
	var results []string
	for _, themeName := range this.getThemeNames() {
		var file = path.Join("themes", themeName, "stylesheets", themeName+".css")
		if _, err := this.StaticAsset(file); err == nil {
			results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s?theme=%s">`, this.JoinStaticURL(file), themeName))
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (this *Context) loadThemeJavaScripts() template.HTML {
	var results []string
	for _, themeName := range this.getThemeNames() {
		for _, ext := range []string{".min", ""} {
			var file = path.Join("themes", themeName, "javascripts", themeName+ext+".js")
			if _, err := this.StaticAsset(file); err == nil {
				results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, this.JoinStaticURL(file), themeName))
				break
			}
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (this *Context) loadAdminJavaScripts() template.HTML {
	var siteName = this.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("custom", strings.ToLower(strings.Replace(siteName, " ", "_", -1)), "js", "admin.js")
	if _, err := this.StaticAsset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, this.JoinStaticURL(file)))
	}
	return ""
}

func (this *Context) loadAdminStyleSheets() template.HTML {
	var siteName = this.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("custom", strings.ToLower(strings.Replace(siteName, " ", "_", -1)), "css", "admin.css")
	if _, err := this.StaticAsset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, this.JoinStaticURL(file)))
	}
	return ""
}

func (this *Context) loadResourceJavaScripts() template.HTML {
	if this.Resource == nil {
		return ""
	}

	var result []string
	for _, file := range []string{"main.js", "locale/" + this.Locale + ".js"} {
		file = path.Join("javascripts", this.Resource.TemplatePath, file)
		if _, err := this.StaticAsset(file); err == nil {
			result = append(result, fmt.Sprintf(`<script src="%s"></script>`, this.JoinStaticURL(file)))
		}
	}
	return template.HTML(strings.Join(result, ""))
}

func (this *Context) loadResourceStyleSheets() template.HTML {
	if this.Resource == nil {
		return ""
	}
	var file = path.Join("stylesheets", this.Resource.TemplatePath, "styles.css")
	if _, err := this.StaticAsset(file); err == nil {
		return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, this.JoinStaticURL(file)))
	}
	return ""
}

func (this *Context) loadActions(action string, subPath ...string) template.HTML {
	var (
		actionKeys     []string
		actionFiles    []assetfsapi.FileInfo
		actions        = map[string]assetfsapi.FileInfo{}
		actionPatterns []assetfs.GlobPatter
		sub            string
	)
	for _, sub = range subPath {
	}

	switch action {
	case "index", "show", "edit", "new":
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", action, sub), TemplateGlob.Wrap("actions", sub)}

		if !this.Resource.Sections.Default.Screen.Show.IsSetI() && action == "edit" {
			actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", "show", sub), TemplateGlob.Wrap("actions", sub)}
		}
	case "global":
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", sub)}
	default:
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", action, sub)}
	}

	var glob = func(pattern assetfs.GlobPatter) {
		this.GlobTemplate(pattern, func(info assetfsapi.FileInfo) {
			actionFiles = append(actionFiles, info)
		})
	}

	for _, pattern := range actionPatterns {
		if this.Anonymous() {
			pattern = pattern.Wrap(AnonymousDirName)
		}
		for _, themeName := range this.getThemeNames() {
			if resourcePath := this.resourcePath(); resourcePath != "" {
				glob(pattern.Wrap("themes", themeName, resourcePath))
			}

			glob(pattern.Wrap("themes", themeName))
		}

		if resourcePath := this.resourcePath(); resourcePath != "" {
			glob(pattern.Wrap(resourcePath))
		}

		glob(pattern)
	}

	// before files have higher priority
	for _, actionFile := range actionFiles {
		actionFileName := strings.TrimSuffix(actionFile.RealPath(), ".tmpl")
		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(path.Base(actionFileName), "")

		if _, ok := actions[base]; !ok {
			actionKeys = append(actionKeys, path.Base(actionFileName))
			actions[base] = actionFile
		}
	}

	sort.Strings(actionKeys)

	result := bytes.NewBufferString("")

	for _, key := range actionKeys {
		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(key, "")
		err := (func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					panic(tracederror.TracedWrap(r, "GetMask error when render action %v", actions[base]))
				}
			}()

			executor, err := this.GetTemplateInfo(actions[base])
			if err == nil {
				err = executor.Execute(result, this, this.FuncValues())
			}
			return
		})()
		if err != nil {
			if et, ok := err.(tracederror.TracedError); ok {
				panic(et)
			}
			result.WriteString(err.Error())
			panic(err)
			return ""
		}
	}

	return template.HTML(strings.TrimSpace(result.String()))
}

func (this *Context) logoutURL() string {
	if this.Admin.Auth != nil {
		if auth.IsAlternated(this.Admin.Auth.Auth(), this.Context) {
			return this.Path(AuthRevertPath)
		}
		return this.Admin.Auth.LogoutURL(this)
	}
	return ""
}

func (this *Context) loginURL() string {
	if this.Admin.Auth != nil {
		return this.Admin.Auth.LoginURL(this)
	}
	return ""
}

func (this *Context) profileURL() string {
	if this.Admin.Auth != nil {
		return this.Admin.Auth.ProfileURL(this)
	}
	return ""
}

func (this *Context) t(key string, defaul ...interface{}) template.HTML {
	return this.T(key, defaul...)
}

func (this *Context) tt(key string, data interface{}, defaul ...interface{}) template.HTML {
	return this.TT(key, data, defaul...)
}

func (this *Context) isSortableMeta(name string) bool {
	return this.Scheme.IsSortableMeta(name)
}

func (this *Context) convertSectionToMetas(sections Sections) []*Meta {
	return sections.ToMetas(func(res *Resource, name string) *Meta {
		return res.GetContextMeta(this, name)
	})
}

func (this *Context) convertSectionToMetasTable(res *Resource, sections Sections) *MetasTable {
	return res.ConvertSectionToMetasTable(sections, res.MetaContextGetter(this))
}

func (this *Context) pageTitle() template.HTML {
	if title := this.Value("page_title"); title != nil {
		switch t := title.(type) {
		case template.HTML:
			return t
		case string:
			return template.HTML(t)
		default:
			return ""
		}
	}
	if this.Action == "search_center" {
		return this.t(I18NGROUP + ".search_center.title")
	}

	if this.Resource == nil {
		if this.PageTitle != "" {
			return this.t(this.PageTitle)
		}
		return template.HTML(this.Admin.Config.DefaultPageTitle(this))
	}

	if this.Action == "action" {
		if action, ok := this.Result.(*Action); ok {
			return this.Resource.GetActionLabel(this, action)
		}
	}

	if this.PageTitle != "" {
		return this.t(this.PageTitle)
	}

	if crumb := this.Breadcrumbs().Last(); crumb != nil {
		if label := crumb.Label(); label == "" {
			return "[NO LABEL]"
		} else if cr, ok := crumb.(*ResourceCrumb); ok {
			if cr.ID != nil {
				return template.HTML(cr.Resource.GetLabel(this, false) + ": " + label)
			}
			return this.t(label)
		} else {
			return this.t(label)
		}
	}

	var (
		defaultValue = this.GetActionLabel()
		titleKey     = fmt.Sprintf(I18NGROUP+".form.%v.title", this.Action)
		usePlural    bool
	)

	if defaultValue == "" {
		defaultValue = "{{.}}"
		if !this.Resource.Config.Singleton {
			usePlural = true
		}
	}

	resourceName := this.Resource.GetLabel(this, usePlural)
	title := fmt.Sprint(this.t(titleKey, defaultValue))

	return utils.RenderHtmlTemplate(title, resourceName)
}

func (this *Context) javaScriptTagSlice(names []string) template.HTML {
	return this.javaScriptTag(names...)
}

func (this *Context) styleSheetTagSlice(names []string) template.HTML {
	return this.styleSheetTag(names...)
}

func (this *Context) globalJavaScriptTag(names ...string) template.HTML {
	var results []string
	prefix := this.Top().JoinStaticURL("javascripts")
	for _, name := range names {
		results = append(results, fmt.Sprintf(`<script src="%s"></script>`, prefix+"/"+name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (this *Context) globalStyleSheetTag(names ...string) template.HTML {
	var results []string
	prefix := this.Top().JoinStaticURL("stylesheets")
	for _, name := range names {
		results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s/%s">`, prefix, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (this *Context) globalJavaScriptTagSlice(names []string) template.HTML {
	return this.globalJavaScriptTag(names...)
}

func (this *Context) globalStyleSheetTagSlice(names []string) template.HTML {
	return this.globalStyleSheetTag(names...)
}
