package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/ecletus/about"
	"github.com/ecletus/helpers"
	"github.com/ecletus/render"
	"github.com/pkg/errors"
	"unapu.com/lib"

	"github.com/jinzhu/inflection"
	oscommon "github.com/moisespsena-go/os-common"
	path_helpers "github.com/moisespsena-go/path-helpers"

	"github.com/moisespsena-go/i18n-modular/i18nmod"

	"github.com/ecletus/common"
	"github.com/moisespsena-go/assetfs"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"

	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
)

// FuncMap return funcs map
func (this *Context) FuncMaps() []funcs.FuncMap {
	top := this.Top()
	if this.jsLibs == nil {
		jsLibFs := this.Admin.Config.AssetFS.NameSpace("static/javascripts/lib")
		this.jsLibs = lib.New(func(dep lib.Dependency) (_ *lib.Lib, err error) {
			var asset assetfs.AssetInterface
			if asset, err = jsLibFs.Asset(dep.String() + "/lib.yml"); err != nil {
				return
			}
			var b []byte
			if b, err = assetfs.Data(asset); err != nil {
				return
			}
			return lib.FromBytes(dep, b)
		})
	}

	funcMaps := []template.FuncMap{
		{
			"js_lib": func(name ...string) (_ string, err error) {
				for _, name := range name {
					if err = this.jsLibs.Add(name); err != nil {
						return
					}
				}
				return "", err
			},
			"js_libs": func() (uris []string, err error) {
				var sorted []*lib.Lib
				if sorted, err = this.jsLibs.Sort(); err != nil {
					return
				}
				uris = make([]string, len(sorted))
				for i, l := range sorted {
					uris[i] = top.JoinStaticURL("javascripts/lib", l.String(), l.Main)
				}
				return
			},
			"qor_context": func() *core.Context { return this.Context },
			"get_route_handler": func() interface{} {
				rctx := this.RouteContext
				return rctx.Handler
			},
			"site":          func() *core.Site { return this.Context.Site },
			"public_url":    func(args ...string) string { return this.Context.Site.PublicURL(args...) },
			"public_urlf":   func(args ...interface{}) string { return this.Context.Site.PublicURLf(args...) },
			"admin_context": func() *Context { return this },
			"current_user": func() common.User {
				return this.CurrentUser()
			},
			"get_resource":         this.Admin.GetResourceByID,
			"new_resource_context": this.NewResourceContext,
			"is_new_record":        this.isNewRecord,
			"is_equal":             this.isEqual,
			"is_included":          this.isIncluded,
			"primary_key_of":       this.primaryKeyOf,
			"unique_key_of":        this.uniqueKeyOf,
			"formatted_value_of":   this.FormattedValueOf,
			"raw_value_of":         this.RawValueOf,
			"result": func() interface{} {
				return this.Result
			},
			"request_uri": func() string {
				return this.Request.RequestURI
			},
			"T":          this.Tt,
			"Ts":         this.TtS,
			"t":          this.t,
			"ts":         this.Ts,
			"tt":         this.tt,
			"flashes":    this.getFlashes,
			"pagination": this.Pagination,
			"escape":     html.EscapeString,
			"raw":        func(str string) template.HTML { return template.HTML(utils.HTMLSanitizer.Sanitize(str)) },
			"unsafe_raw": func(str string) template.HTML { return template.HTML(str) },
			"equal":      equal,
			"stringify":  this.Stringify,
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

			"yield": func(s *template.State) string {
				this.Yield(s.Writer(), this.Result)
				return ""
			},
			"include": func(s *template.State, name string, result ...interface{}) string {
				this.Include(s.Writer(), name, result...)
				return ""
			},
			"render":      this.RenderHtml,
			"render_text": this.renderText,
			"render_with": this.renderWith,
			"render_form": this.renderForm,
			"render_meta": func(state *template.State, value interface{}, meta *Meta, types ...string) {
				var (
					typ = "index"
				)

				for _, t := range types {
					typ = t
				}

				this.renderMeta(state, meta, value, []string{}, typ, NewTrimLeftWriter(state.Writer()))
			},
			"render_filter": this.renderFilter,
			"saved_filters": this.savedFilters,
			"has_filter": func() bool {
				query := this.Request.URL.Query()
				for key := range query {
					if regexp.MustCompile("filter[(\\w+)]").MatchString(key) && query.Get(key) != "" {
						return true
					}
				}
				return false
			},
			"page_title": this.pageTitle,
			"meta_label": func(meta *Meta) template.HTML {
				return template.HTML(meta.GetLabelC(this.Context))
			},
			"meta_help": func(meta *Meta) template.HTML {
				key, defaul := meta.GetHelpPair()
				if key != "" {
					defaul = strings.TrimSpace(this.Admin.Ts(this.Context, key, defaul))
				}
				if strings.Contains(defaul, "{{") {
					tmpl, err := template.New("meta_help{" + meta.Name + "}").Parse(defaul)
					if err != nil {
						return template.HTML("[[parse template failed: " + err.Error() + "]]")
					}
					var w bytes.Buffer
					if err = this.ExecuteTemplate(tmpl, &w, this); err != nil {
						return template.HTML("[[execute template failed: " + err.Error() + "]]")
					}
					return template.HTML(w.String())
				}
				return template.HTML(defaul)
			},
			"meta_record_label": func(meta *Meta, record interface{}) template.HTML {
				key, defaul := meta.GetRecordLabelPair(this, record)
				if key != "" {
					return template.HTML(strings.TrimSpace(this.Admin.Ts(this.Context, key, defaul)))
				}
				return template.HTML(defaul)
			},
			"meta_record_help": func(meta *Meta, record interface{}) template.HTML {
				var key, defaul string
				if this.Type.Has(SHOW) {
					key, defaul = meta.GetRecordShowHelpPair(this, record)
				} else {
					key, defaul = meta.GetRecordHelpPair(this, record)
				}
				if key != "" {
					defaul = strings.TrimSpace(this.Admin.Ts(this.Context, key, defaul))
				}
				if strings.Contains(defaul, "{{") {
					tmpl, err := template.New("meta_help{" + meta.Name + "}").Parse(defaul)
					if err != nil {
						return template.HTML("[[parse template failed: " + err.Error() + "]]")
					}
					var w bytes.Buffer
					if err = this.ExecuteTemplate(tmpl, &w, this); err != nil {
						return template.HTML("[[execute template failed: " + err.Error() + "]]")
					}
					return template.HTML(w.String())
				}
				return template.HTML(defaul)
			},
			"section_title": func(section *Section) (s template.HTML) {
				key := section.Resource.I18nPrefix + ".sections." + section.Title
				return template.HTML(strings.TrimSpace(this.Admin.Ts(this.Context, key, section.Title)))
			},
			"section_help": func(section *Section, readOnly bool) (s template.HTML) {
				key := section.Resource.I18nPrefix + ".sections." + section.Title + "_"
				defaul := section.Help
				if readOnly {
					key += "ro_help"
					defaul = section.ReadOnlyHelp
				} else {
					key += "help"
				}
				return template.HTML(strings.TrimSpace(this.Admin.Ts(this.Context, key, defaul)))
			},
			"resource_help": func(res *Resource) template.HTML {
				key, defaul := res.GetHelpPair()
				if key != "" {
					return this.Admin.T(this.Context, key, defaul)
				}
				return ""
			},
			"resource_plural_help": func(res *Resource) template.HTML {
				key, defaul := res.GetPluralHelpPair()
				if key != "" {
					return this.Admin.T(this.Context, key, defaul)
				}
				return ""
			},
			"meta_placeholder": func(meta *Meta, context *Context, placeholder string) template.HTML {
				if getPlaceholder, ok := meta.Config.(interface {
					GetPlaceholder(*Context) (template.HTML, bool)
				}); ok {
					if str, ok := getPlaceholder.GetPlaceholder(context); ok {
						return str
					}
				}

				key := fmt.Sprintf("%v.attributes.%v.placeholder", meta.BaseResource.I18nPrefix, meta.Name)
				return context.Admin.T(context.Context, key, func(context i18nmod.Context) *i18nmod.T {
					var key string
					if meta.Config != nil {
						key = i18nmod.PkgToGroup(path_helpers.PkgPathOf(meta)) + ".metas." + meta.Type + ".placeholder"
					} else {
						key = I18NGROUP + ".metas." + meta.Type + ".placeholder"
					}
					return context.T(key).Default(placeholder)
				})
			},

			"url_for":            this.URLFor,
			"top_url_for":        this.TopURLFor,
			"link_to":            this.linkTo,
			"link_to_ajax_load":  this.linkToAjaxLoad,
			"patch_current_url":  this.PatchCurrentURL,
			"patch_url":          this.PatchURL,
			"join_current_url":   this.JoinCurrentURL,
			"join_url":           this.JoinURL,
			"logout_url":         this.logoutURL,
			"login_url":          this.loginURL,
			"profile_url":        this.profileURL,
			"search_center_path": func() string { return this.JoinPath("!search") },
			"new_resource_path":  this.newResourcePath,
			"defined_resource_show_page": func(res *Resource) bool {
				if res != nil {
					if res.Top().isSetShowAttrs {
						return true
					}
				}

				return false
			},
			"get_menus":                 this.getMenus,
			"get_resource_menus":        this.getResourceMenus,
			"get_resource_item_menus":   this.getResourceItemMenus,
			"get_resource_menu_actions": this.getResourceMenuActions,
			"get_scopes":                this.GetScopes,
			"get_formatted_errors":      this.getFormattedErrors,
			"load_actions":              this.loadActions,
			"allowed_actions":           this.AllowedActions,
			"is_sortable_meta":          this.isSortableMeta,
			"index_sections":            this.indexSections,
			"show_sections":             this.showSections,
			"new_sections":              this.newSections,
			"edit_sections":             this.editSections,
			"show_meta_sections":        this.showMetaSections,
			"new_meta_sections":         this.newMetaSections,
			"edit_meta_sections":        this.editMetaSections,
			"convert_sections_to_metas": this.convertSectionToMetas,

			"has_create_permission": this.hasCreatePermission,
			"has_read_permission":   this.hasReadPermission,
			"has_update_permission": this.hasUpdatePermission,
			"has_delete_permission": this.hasDeletePermission,

			"read_permission_filter": this.readPermissionFilter,

			"qor_theme_class": this.themesClass,

			"javascript_tag":              this.javaScriptTag,
			"javascript_tag_slice":        this.javaScriptTagSlice,
			"global_javascript_tag":       this.globalJavaScriptTag,
			"global_javascript_tag_slice": this.globalJavaScriptTagSlice,

			"stylesheet_tag":              this.styleSheetTag,
			"stylesheet_tag_slice":        this.styleSheetTagSlice,
			"global_stylesheet_tag":       this.globalStyleSheetTag,
			"global_stylesheet_tag_slice": this.globalStyleSheetTagSlice,

			"load_theme_stylesheets":    this.loadThemeStyleSheets,
			"load_theme_javascripts":    this.loadThemeJavaScripts,
			"load_admin_stylesheets":    this.loadAdminStyleSheets,
			"load_admin_javascripts":    this.loadAdminJavaScripts,
			"load_resource_stylesheets": this.loadResourceStyleSheets,
			"load_resource_javascripts": this.loadResourceJavaScripts,
			"load_print_mode_stylesheeets": func() template.HTML {
				if _, ok := this.Request.URL.Query()["print"]; ok {
					return this.styleSheetTag("print")
				}
				return ""
			},

			"global_url": top.Path,
			"static_url": top.JoinStaticURL,
			"url":        this.Path,
			"admin_public_url": func(p ...string) string {
				if this.Admin.Config.MountPath == "/" {
					return this.Site.PublicURL(p...)
				}
				return this.Site.PublicURL(append([]string{strings.Trim(this.Admin.Config.MountPath, "/")}, p...)...)
			},
			"admin_static_url": this.JoinStaticURL,
			"locale": func() string {
				return this.Locale
			},
			"time_loc": func() *time.Location {
				return this.TimeLocation
			},
			"crumbs": func() []core.Breadcrumb {
				return this.Breadcrumbs().ItemsWithoutLast()
			},
			"resource_key": func() aorm.ID {
				return this.ResourceID
			},
			"resource_parent_keys": func() []aorm.ID {
				return this.ParentResourceID
			},
			"new_resource_struct": func(res *Resource) interface{} {
				return res.NewStruct(this.Site)
			},
			"is_nil": helpers.IsNilInterface,

			"htmlify":       this.Htmlify,
			"htmlify_items": this.HtmlifyItems,
			"has_prefix": func(str interface{}, prefix string) bool {
				return strings.HasPrefix(fmt.Sprint(str), prefix)
			},
			"has_sufix": func(str interface{}, sufix string) bool {
				return strings.HasSuffix(fmt.Sprint(str), sufix)
			},

			"sprintf": fmt.Sprintf,

			"alert": func(arg template.HTML) interface{} {
				if strings.TrimSpace(string(arg)) != "" {
					this.Alerts = append(this.Alerts, arg)
				}
				return nil
			},
			"alerts": func() []template.HTML {
				return this.Alerts
			},
			"admin_i18n": func(key string) string {
				return I18NGROUP + key
			},
			"b64": func(value interface{}) string {
				switch t := value.(type) {
				case []byte:
					return base64.StdEncoding.EncodeToString(t)
				case string:
					return base64.StdEncoding.EncodeToString([]byte(t))
				default:
					return ""
				}
			},
			"about": func() about.Abouter {
				return this.Admin.Config.SiteAbouter(this)
			},

			"record_frame": func(s *template.State, res *Resource, record interface{}, name string) {
				if renderer := res.GetFrameRenderer(name); renderer != nil {
					if err := renderer.Render(this, s); err != nil {
						panic(errors.Wrapf(err, "record frame renderer of %q", name))
					}
					return
				}
				templateNames := res.GetFrameRendererTemplateName(this, name)
				if this.Anonymous() {
					var newTemplateNames []string
					for _, templateName := range templateNames {
						newTemplateNames = append(newTemplateNames, path.Join(path.Dir(templateName), AnonymousDirName, path.Base(templateName)), templateName)
					}
					templateNames = newTemplateNames
				}
				for _, templateName := range templateNames {
					if executor, err := this.GetTemplate(templateName); err == nil {
						defer this.WithResult(record)()
						if err = executor.Execute(s.Writer(), this); err != nil {
							panic(errors.Wrapf(err, "record frame render of %q", name))
						}
						return
					} else if !oscommon.IsNotFound(err) {
						panic(errors.Wrapf(err, "record frame render of %q", name))
					}
				}
			},

			"is_show": func() bool {
				return this.Type.Has(SHOW)
			},
			"not_show": func() bool {
				return !this.Type.Has(SHOW)
			},
			"default": func(values ...interface{}) interface{} {
				for _, v := range values {
					if !aorm.IsBlank(reflect.Indirect(reflect.ValueOf(v))) {
						return v
					}
				}
				return nil
			},
			"must_config_get": func(configor core.Configor, key interface{}) (v interface{}) {
				v, _ = configor.ConfigGet(key)
				return v
			},
			"render_scripts": func(s *template.State) (r template.HTML) {
				var (
					c = this.Context
					w bytes.Buffer
				)

				for _, h := range []render.ScriptHandlers{this.PageHandlers.ScriptHandlers, render.GetScriptHandlers(s.Context())} {
					for _, h := range h {
						if err := h.Handler(s, c, &w); err != nil {
							w.WriteString("[[render execute script handler `" + h.Name + "` failed: " + err.Error() + "]]")
							break
						}
					}
				}
				return template.HTML(w.String())
			},

			"render_styles": func(s *template.State) (r template.HTML) {
				var (
					c = this.Context
					w bytes.Buffer
				)

				for _, h := range []render.StyleHandlers{this.PageHandlers.StyleHandlers, render.GetStyleHandlers(s.Context())} {
					for _, h := range h {
						if err := h.Handler(s, c, &w); err != nil {
							w.WriteString("[[render execute style handler `" + h.Name + "` failed: " + err.Error() + "]]")
							break
						}
					}
				}
				return template.HTML(w.String())
			},

			"form": func(s *template.State, name string, pipes ...interface{}) template.HTML {
				var c = this.Context
				state := &render.FormState{name, s.Exec(name, pipes...)}
				for _, h := range this.PageHandlers.FormHandlers.AppendCopy(render.GetFormHandlers(s.Context())...) {
					if err := h.Handler(state, c); err != nil {
						return template.HTML("[[render execute form handler `" + h.Name + "` for `" + name + "` form failed: " + err.Error() + "]]")
					}
				}
				return template.HTML(state.Body)
			},

			"media_url": func(pth string, storageName ...string) string {
				if len(storageName) == 0 {
					return this.MediaURL("default", pth)
				}
				return this.MediaURL(storageName[0], pth)
			},
		},
		this.Admin.funcMaps,
	}

	funcMaps = append(funcMaps, this.funcMaps...)

	return funcMaps
}
