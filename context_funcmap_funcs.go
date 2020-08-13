package admin

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/tracederror"
	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
	"github.com/pkg/errors"
)

var TemplateExecutorMetaValue = template.Must(template.New(PKG + ".meta_value").Parse("{{.Value}}")).CreateExecutor()

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
	return res.GetContextIndexURI(this.Context) + "/new"
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
				result = reflectValue.Elem().Interface()
			}

			if meta.Type == "number" || meta.Type == "float" {
				if this.isNewRecord(value) && equal(reflect.Zero(reflect.TypeOf(result)).Interface(), result) {
					return nil
				}
			} else if ID, ok := result.(aorm.ID); ok {
				return ID.String()
			}
			return originalResult
		}
		return nil
	}

	if !meta.Virtual {
		panic(fmt.Errorf("No valuer found for meta %v of resource %v", meta.Name, meta.BaseResource.Name))
	}
	return nil
}

func (this *Context) renderForm(state *template.State, value interface{}, sections []*Section) {
	if len(*this.metaPath) == 0 {
		*this.metaPath = append(*this.metaPath, "QorResource")
		defer func() {
			*this.metaPath = (*this.metaPath)[0 : len(*this.metaPath)-1]
		}()
	}
	this.renderSections(state, value, sections, []string{"QorResource"}, state.Writer(), "form", this.Type.Has(SHOW) || this.Type.Has(INDEX))
}

func (this *Context) renderSections(state *template.State, value interface{}, sections []*Section, prefix []string, writer io.Writer, kind string, readOnly bool) {
	var (
		res     *Resource
		getMeta func(string) *Meta
	)
	for _, section := range sections {
		var (
			hasRequired bool
			rows        []struct {
				Length      int
				ColumnsHTML template.HTML
			}
		)

		if res != section.Resource {
			res = section.Resource
			getMeta = section.Resource.MetaContextGetter(this)
		}

		for _, column := range section.Rows {
			var (
				columnsHTML bytes.Buffer
				w           = NewTrimLeftWriter(&columnsHTML)
				exclude     int
			)
			for _, col := range column {
				meta := getMeta(col)
				if meta != nil {
					if meta.Enabled == nil || meta.Enabled(value, this, meta) {
						if meta.IsRequired() {
							hasRequired = true
						}
						this.renderMeta(state, meta, value, prefix, kind, w)
					} else {
						exclude++
					}
				}
			}

			if !hasRequired && (this.Action == "show" || this.Action == "action_show") && w.Empty() {
				continue
			}

			rows = append(rows, struct {
				Length      int
				ColumnsHTML template.HTML
			}{
				Length:      len(column) - exclude,
				ColumnsHTML: template.HTML(columnsHTML.Bytes()),
			})
		}

		if len(rows) > 0 {
			var data = map[string]interface{}{
				"Section":  section,
				"Title":    template.HTML(section.Title),
				"Rows":     rows,
				"ReadOnly": readOnly,
			}

			if executor, err := this.GetTemplate("metas/section"); err == nil {
				err = executor.Execute(writer, data, this.FuncValues())
			}
		}
	}
}

func (this *Context) renderFilter(filter *Filter) template.HTML {
	var (
		err      error
		executor *template.Executor
		dir      = "filter"
		result   = bytes.NewBufferString("")
	)

	if filter.advanced {
		dir = "advanced_filter"
	}

	if executor, err = this.GetTemplate(fmt.Sprintf("metas/%v/%v", dir, filter.Type)); err == nil {
		var label string
		if !filter.LabelDisabled {
			label = filter.GetLabelC(this.Context)
		}
		var data = map[string]interface{}{
			"Filter":          filter,
			"Label":           label,
			"InputNamePrefix": fmt.Sprintf("filtersByName[%v]", filter.Name),
			"Context":         this,
			"Resource":        this.Resource,
		}

		err = executor.Execute(result, data, this.FuncValues())
	}

	if err != nil {
		this.AddError(err)
		result.WriteString(errors.Wrap(err, fmt.Sprintf("render filter template for %v(%v)", filter.Name, filter.Type)).Error())
	}

	return template.HTML(result.String())
}

func (this *Context) savedFilters() (filters []SavedFilter) {
	this.Admin.settings.Get("saved_filters", &filters, this)
	return
}

func (this *Context) NestedForm() bool {
	return this.nestedForm > 0
}

func (this *Context) renderMeta(state *template.State, meta *Meta, record interface{}, prefix []string, metaType string, writer io.Writer) {
	var (
		err             error
		funcsMap        = funcs.FuncMap{}
		executor        *template.Executor
		formattedValue  = this.FormattedValueOf(record, meta)
		show            = this.Type.Has(SHOW) || this.Type.Has(INDEX)
		nestedFormCount int
		readOnly        = show
	)
	*this.metaPath = append(*this.metaPath, meta.Name)
	defer func() {
		*this.metaPath = (*this.metaPath)[0 : len(*this.metaPath)-1]
	}()

	if show && !meta.IsRequired() {
		if !meta.ForceShowZero && meta.IsZero(record, formattedValue) {
			return
		}
		if !meta.ForceEmptyFormattedRender {
			if formattedValue == nil {
				return
			}
			switch t := formattedValue.(type) {
			case string:
				if len(t) == 0 {
					return
				}
			case template.HTML:
				if len(t) == 0 {
					return
				}
			case aorm.Zeroer:
				if t.IsZero() {
					return
				}
			}
		}
	}

	if !meta.Include {
		prefix = append(prefix, meta.Name)
	}

	var generateNestedRenderSections = func(kind string) func(*template.State, interface{}, *Meta, []*Section, int, ...string) {
		return func(state *template.State, record interface{}, meta *Meta, sections []*Section, index int, prefx ...string) {
			this.nestedForm++
			if index == -2 {
				*this.metaPath = append(*this.metaPath, "{{index}}")
			} else {
				*this.metaPath = append(*this.metaPath, strconv.Itoa(nestedFormCount))
			}
			defer func() {
				nestedFormCount++
				this.nestedForm--
				*this.metaPath = (*this.metaPath)[0 : len(*this.metaPath)-1]
			}()

			newPrefix := append([]string{}, prefix...)

			if len(prefx) > 0 && prefx[0] != "" {
				for prefx[0][0] == '.' {
					newPrefix = newPrefix[0 : len(newPrefix)-1]
					prefx[0] = prefx[0][1:]
				}

				newPrefix = append(newPrefix, prefx...)
			}

			if index >= 0 {
				last := newPrefix[len(newPrefix)-1]
				newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index))
			} else if index == -2 {
				last := newPrefix[len(newPrefix)-1]
				newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[{{index}}]", last))
			}

			if len(sections) > 0 {
				w := NewTrimLeftWriter(state.Writer())
				w.Before = func() {
					if this.Type.Has(EDIT) {
						if !sections[0].Resource.Config.DisableFormID {
							for _, field := range aorm.StructOf(record).PrimaryFields {
								if meta := sections[0].Resource.GetMeta(field.Name); meta != nil {
									this.renderMeta(state, meta, record, newPrefix, kind, w)
								}
							}
						}
					}
				}
				this.renderSections(state, record, sections, newPrefix, w, kind, readOnly)
			}
		}
	}

	funcsMap["render_nested_form"] = generateNestedRenderSections("form")

	defer func() {
		if err != nil {
			panic(err)
		}
		if r := recover(); r != nil {
			var msg string
			msg = fmt.Sprintf("render meta %q (%v)", meta.Name, meta.Type)
			writer.Write([]byte(msg))

			if et, ok := r.(tracederror.TracedError); ok {
				panic(tracederror.Wrap(et, msg))
			} else if err, ok := r.(error); ok {
				panic(tracederror.New(errors.Wrap(err, msg), et.Trace()))
			} else {
				panic(tracederror.New(errors.Wrap(fmt.Errorf("recoverd_error %T: %v", r, r), msg)))
			}
		}
	}()

	switch {
	case meta.Config != nil:
		if templater, ok := meta.Config.(interface {
			GetTemplate(context *Context, metaType string) (*template.Executor, error)
		}); ok {
			if executor, err = templater.GetTemplate(this, metaType); err == nil {
				break
			}
		}
		fallthrough
	default:
		var others []string
		metaUserType := meta.GetType(record, this)

		if metaUserType != "" {
			others = append(others, fmt.Sprintf("metas/%v/%v", metaType, metaUserType))
		}

		if executor, err = this.GetTemplateOrDefault(fmt.Sprintf("%v/metas/%v/%v", meta.BaseResource.ToParam(), metaType, meta.Name),
			TemplateExecutorMetaValue, others...); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("haven't found %v template for meta %v", metaType, meta.Name))
		}
	}

	if err == nil {
		if !readOnly {
			readOnly = meta.IsReadOnly(this, record)
		}
		var data = map[string]interface{}{
			"Context":       this,
			"BaseResource":  meta.BaseResource,
			"Meta":          meta,
			"Record":        record,
			"ResourceValue": record,
			"Value":         formattedValue,
			"Label":         meta.Label,
			"InputName":     strings.Join(prefix, "."),
			"ReadOnly":      readOnly,
			"NotReadOnly":   !readOnly,
		}
		data["InputId"] = strings.Join(*this.metaPath, "_")
		executor.SetSuper(state)
		err = executor.Execute(writer, data, this.FuncValues(), funcsMap)
	}

	if err != nil {
		err = tracederror.TracedWrap(err, "got error when render meta %v template for %v(%v)", metaType, meta.Name, meta.Type)
	}
}

func (this *Context) isEqual(value interface{}, hasValue interface{}) bool {
	if (value == nil && hasValue != nil) || (value != nil && hasValue == nil) {
		return false
	}

	var result string

	if reflect.Indirect(reflect.ValueOf(hasValue)).Kind() == reflect.Struct {
		result = aorm.IdOf(hasValue).String()
	} else {
		result = fmt.Sprint(hasValue)
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.Kind() == reflect.Struct {
		return aorm.IdOf(hasValue).String() == result
	}

	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
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

func (this *Context) indexSections(resources ...*Resource) []*Section {
	res := this.getResource(resources...)
	if this.Layout != "" && this.Layout != "index" {
		layout := res.GetAdminLayout(this.Layout)
		attrs := layout.Metas
		if layout.NotIndexRenderID && !this.Api {
			attrs = append(attrs, "-"+layout.MetaID)
		}
		sections := res.SectionsList(attrs)
		sections = res.allowedSections(nil, sections, this, roles.Read)
		return sections
	}
	scheme := this.Scheme
	if scheme == nil {
		scheme = res.Scheme
	}
	sections := scheme.IndexSections(this)
	return sections
}

func (this *Context) editSections(res *Resource, recorde ...interface{}) []*Section {
	if len(recorde) == 0 {
		recorde = append(recorde, nil)
	}
	return res.EditSections(this, recorde[0])
}

func (this *Context) newSections(resources ...*Resource) []*Section {
	res := this.getResource(resources...)
	return res.NewSections(this)
}

func (this *Context) showSections(recorde interface{}, resources ...*Resource) []*Section {
	res := this.getResource(resources...)
	return res.ShowSections(this, recorde)
}

func (this *Context) editMetaSections(meta *Meta, record interface{}) []*Section {
	if meta.Resource == nil {
		panic("admin.func_map.editMetaSections: meta.Resource is nil")
	}
	if meta.Resource.ModelStruct.Parent == nil {
		this = this.Clone()
		_, this.Context = this.Context.NewChild(nil)
		this.Resource = meta.Resource
		this.Result = record
	} else {
		this = this.CreateChild(meta.Resource, record)
	}
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.EditSections(this, record)
		}
	}
	if len(attrs) == 0 {
		attrs = res.EditAttrs()
	}
	secs := res.allowedSections(record, attrs, this, roles.Update)
	return secs
}

func (this *Context) newMetaSections(meta *Meta, record interface{}) []*Section {
	if meta.Resource == nil {
		panic("admin.func_map.newMetaSections: meta.Resource is nil")
	}
	this = this.CreateChild(meta.Resource, record)
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.NewSections(this)
		}
	}
	if len(attrs) == 0 {
		attrs = res.NewAttrs()
	}
	return res.allowedSections(record, attrs, this, roles.Create)
}

func (this *Context) showMetaSections(meta *Meta, record interface{}) []*Section {
	if meta.Resource == nil {
		panic("admin.func_map.showMetaSections: meta.Resource is nil")
	}
	this = this.CreateChild(meta.Resource, record)
	res := meta.Resource
	var attrs []*Section
	if meta.Config != nil {
		if sc, ok := meta.Config.(*SingleEditConfig); ok {
			attrs = sc.ShowSections(this, record)
		}
	}
	if len(attrs) == 0 {
		attrs = res.ShowSections(this, record)
	}
	return res.allowedSections(record, attrs, this, roles.Read)
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

func (this *Context) loadActions(action string) template.HTML {
	var (
		actionKeys     []string
		actionFiles    []assetfsapi.FileInfo
		actions        = map[string]assetfsapi.FileInfo{}
		actionPatterns []assetfs.GlobPatter
	)

	switch action {
	case "index", "show", "edit", "new":
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", action), TemplateGlob.Wrap("actions")}

		if !this.Resource.isSetShowAttrs && action == "edit" {
			actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", "show"), TemplateGlob.Wrap("actions")}
		}
	case "global":
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions")}
	default:
		actionPatterns = []assetfs.GlobPatter{TemplateGlob.Wrap("actions", action)}
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
			return template.HTML("")
		}
	}

	return template.HTML(strings.TrimSpace(result.String()))
}

func (this *Context) logoutURL() string {
	if this.Admin.Auth != nil {
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

func (this *Context) isSortableMeta(meta *Meta) bool {
	for _, attr := range this.Resource.SortableAttrs() {
		if attr == meta.Name && meta.FieldStruct != nil && meta.FieldStruct.IsNormal && meta.FieldStruct.DBName != "" {
			return true
		}
	}
	return false
}

func (this *Context) convertSectionToMetas(res *Resource, sections []*Section) []*Meta {
	return res.ConvertSectionToMetas(sections, res.MetaContextGetter(this))
}

type formattedError struct {
	Label  string
	Errors []string
}

func (this *Context) getFormattedErrors() (formatedErrors []formattedError) {
	type labelInterface interface {
		Label() string
	}
	ctx := this.GetI18nContext()

	for _, err := range this.GetErrors() {
		if labelErr, ok := err.(labelInterface); ok {
			var found bool
			label := labelErr.Label()
			for _, formatedError := range formatedErrors {
				if formatedError.Label == label {
					formatedError.Errors = append(formatedError.Errors, core.StringifyErrorT(ctx, err))
					found = true
				}
			}
			if !found {
				formatedErrors = append(formatedErrors, formattedError{Label: label, Errors: []string{core.StringifyErrorT(ctx, err)}})
			}
		} else {
			formatedErrors = append(formatedErrors, formattedError{Errors: []string{core.StringifyErrorT(ctx, err)}})
		}
	}
	return
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
