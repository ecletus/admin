package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"path"
	"reflect"
	"strings"

	"github.com/ecletus/about"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils/url"
	"github.com/ecletus/helpers"
	"github.com/ecletus/render"
	"github.com/moisespsena-go/maps"
	"github.com/pkg/errors"
	"unapu.com/lib"

	"github.com/jinzhu/inflection"
	oscommon "github.com/moisespsena-go/os-common"
	path_helpers "github.com/moisespsena-go/path-helpers"

	"github.com/ecletus/auth"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/i18n-modular/i18nmod"

	"github.com/ecletus/common"
	"github.com/moisespsena-go/assetfs"

	"github.com/ecletus/core"
	"github.com/ecletus/core/utils"
	"github.com/moisespsena-go/aorm"

	"github.com/moisespsena/template/funcs"
	"github.com/moisespsena/template/html/template"
)

func hasPermission(defaul *Context, do func(this *Context, perm Permissioner) bool) func(arg interface{}, args ...interface{}) bool {
	return func(arg interface{}, args ...interface{}) bool {
		this := defaul
		switch t := arg.(type) {
		case *Context:
			this = t
			arg = args[0]
		}
		switch t := arg.(type) {
		case *Resource:
			return do(this, t)
		case *Meta:
			return do(this, t)
		case core.Permissioner:
			return do(this, NewPermissioner(func(mode roles.PermissionMode, ctx *Context) (perm roles.Perm) {
				return t.HasPermission(mode, ctx.Context)
			}))
		default:
			return do(this, arg.(Permissioner))
		}
	}
}

func hasRecordPermission(defaul *Context, do func(this *Context) bool) func(args ...interface{}) (bool, error) {
	return func(args ...interface{}) (bool, error) {
		switch len(args) {
		case 0:
			return do(defaul), nil
		case 1:
			return do(args[0].(*Context)), nil
		case 2:
			return do(defaul.CreateChild(args[0].(*Resource), args[1])), nil
		default:
			return false, fmt.Errorf("wrong number of args for admin.Context.hasRecordPermission: want at least 0 or 1 or 2, but got %d", len(args))
		}
	}
}

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

	renderMeta := func(state *template.State, this *Context, value interface{}, meta *Meta, types ...string) {
		var (
			typ  = "index"
			mode string
		)

		for _, t := range types {
			if strings.HasPrefix(t, "mode-") {
				mode = strings.TrimPrefix(t, "mode-")
			} else if t != "" {
				typ = t
			}
		}

		this.renderMeta(state, meta, value, []string{}, typ, mode, NewTrimLeftWriter(state.Writer()))
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
			"make_opts": func(args ...interface{}) maps.Map {
				if len(args) == 0 {
					return nil
				}
				m := maps.Map{}
				if len(args)%2 == 0 {
					for i := 0; i < len(args); i += 2 {
						m[args[i]] = args[i+1]
					}
				}
				return m
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
				if this.Yield == nil {
					this.Yielder(s.Writer(), this.Result)
				} else {
					this.Yield(s.Writer(), this.Result)
				}
				return ""
			},
			"include": func(s *template.State, name string, result ...interface{}) string {
				this.Include(s.Writer(), name, result...)
				return ""
			},
			"include_record": func(s *template.State, name string, record interface{}) string {
				this.IncludeRecord(s.Writer(), name, record)
				return ""
			},
			"render":      this.RenderHtml,
			"render_text": this.renderText,
			"render_with": this.renderWith,
			"render_form": this.renderForm,
			"render_meta": func(state *template.State, value interface{}, meta *Meta, types ...string) {
				renderMeta(state, this, value, meta, types...)
			},
			"render_meta_ctx":       renderMeta,
			"render_meta_with_path": this.renderMetaWithPath,
			"render_meta_with_path_ctx": func(state *template.State, ctx *Context, pth string, meta *Meta, types ...string) {
				ctx.renderMetaWithPath(state, pth, ctx.Result, meta, types...)
			},
			"render_filter": this.renderFilter,
			"saved_filters": this.savedFilters,
			"has_filter": func() bool {
				query := this.Request.URL.Query()
				for key := range query {
					if name, _ := GetFilterFromQS(key); name != "" && query.Get(key) != "" {
						return true
					}
				}
				return false
			},
			"requested_filters_and_scopes": func() (res struct {
				Filters map[string][]struct{ Name, Value string }
				Scopes  []string
			}) {
				var (
					query = this.Request.URL.Query()
				)
				res.Filters = map[string][]struct{ Name, Value string }{}

				for key := range query {
					if name, fKey := GetFilterFromQS(key); name != "" {
						if value := query.Get(key); value != "" {
							if _, ok := res.Filters[name]; !ok {
								res.Filters[name] = nil
							}
							res.Filters[name] = append(res.Filters[name], struct{ Name, Value string }{Name: fKey, Value: value})
						}
					}
				}
				res.Scopes = query["scope[]"]
				return
			},
			"page_title": this.pageTitle,
			"meta_label": func(meta *Meta) template.HTML {
				return template.HTML(meta.GetLabelC(this.Context))
			},
			"meta_help": func(meta *Meta) template.HTML {
				return template.HTML(meta.GetHelp(this))
			},
			"meta_help_ctx": func(this *Context, meta *Meta) template.HTML {
				return template.HTML(meta.GetHelp(this))
			},
			"meta_record_label": func(meta *Meta, record interface{}) template.HTML {
				return template.HTML(meta.GetRecordLabel(this, record))
			},
			"meta_record_help": func(meta *Meta, record interface{}, ro ...bool) template.HTML {
				if (len(ro) > 0 && ro[0]) || this.ReadOnly || this.Type.Has(SHOW) {
					return template.HTML(meta.GetRecordShowHelp(this, record))
				}
				return template.HTML(meta.GetRecordHelp(this, record))
			},
			"table_header_title": func(h *MetaTableHeader) template.HTML {
				if h.Section != nil {
					key := h.Section.Resource.I18nPrefix + ".sections." + h.Section.Title
					return template.HTML(strings.TrimSpace(this.Admin.Ts(this.Context, key, h.Section.Title)))
				}
				return template.HTML(h.Meta.GetLabelC(this.Context))
			},
			"section_title": func(s interface{}) template.HTML {
				var section *Section
				switch t := s.(type) {
				case *Section:
					section = t
				case *TreeSection:
					section = t.Section
				case *MetaTableHeader:
					section = t.Section
				}
				if section.Title != "" {
					key := section.Resource.I18nPrefix + ".sections." + section.Title
					return template.HTML(strings.TrimSpace(this.Admin.Ts(this.Context, key, section.Title)))
				}
				return ""
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

			"url_for":           this.URLFor,
			"top_url_for":       this.TopURLFor,
			"link_to":           this.linkTo,
			"link_to_ajax_load": this.linkToAjaxLoad,
			"patch_current_url": this.PatchCurrentURL,
			"patch_url":         this.PatchURL,
			"join_current_url":  this.JoinCurrentURL,
			"join_url":          this.JoinURL,
			"url_param": func(name string, values ...interface{}) *url.Param {
				var p = &url.Param{Name: name}
				for _, v := range values {
					p.Values = append(p.Values, fmt.Sprint(v))
				}
				return p
			},
			"url_flag": func(name string, value bool) *url.FlagParam {
				return &url.FlagParam{Name: name, Value: value}
			},
			"logout_url":         this.logoutURL,
			"login_url":          this.loginURL,
			"profile_url":        this.profileURL,
			"search_center_path": func() string { return this.JoinPath("!search") },
			"new_resource_path":  this.newResourcePath,
			"defined_resource_show_page": func(res *Resource) bool {
				if res != nil {
					if res.Top().Scheme.Sections.Default.Screen.Show.IsSetI() {
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
			"get_formatted_errors": func() []core.FormattedError {
				return append(this.GetCleanFormattedErrors(), this.GetCleanFormattedErrorsOf(&this.Warnings)...)
			},
			"load_actions":    this.loadActions,
			"allowed_actions": this.AllowedActions,

			"index_sections": func(...interface{}) {
				panic("deprecated. uses 'index_sections_ctx'")
			},
			"index_sections_ctx": func(this *Context) []*Section {
				return this.indexSections()
			},
			"show_sections": func(...interface{}) {
				panic("deprecated. uses 'show_sections_ctx'")
			},
			"show_sections_ctx": func(this *Context) []*Section {
				return this.showSections()
			},
			"new_sections": func(...interface{}) {
				panic("deprecated. uses 'new_sections_ctx'")
			},
			"new_sections_ctx": func(this *Context, res *Resource) []*Section {
				return this.newSections(res)
			},
			"edit_sections": this.editSections,
			"edit_sections_ctx": func(this *Context, res *Resource, rec ...interface{}) []*Section {
				return this.editSections(res, rec...)
			},
			"show_meta_sections": func(...interface{}) {
				panic("deprecated. uses 'show_meta_sections_ctx'")
			},
			"show_meta_sections_ctx": func(this *Context, meta *Meta) []*Section {
				return this.showMetaSections(meta)
			},
			"new_meta_sections": func(...interface{}) {
				panic("deprecated. uses 'new_meta_sections_ctx'")
			},
			"new_meta_sections_ctx": func(this *Context, meta *Meta) []*Section {
				return this.newMetaSections(meta)
			},
			"edit_meta_sections": func(...interface{}) {
				panic("deprecated. uses 'edit_meta_sections_ctx'")
			},
			"edit_meta_sections_ctx": func(self *Context, meta *Meta) []*Section {
				_ = this
				return self.editMetaSections(meta)
			},
			"convert_sections_to_metas": func(...interface{}) {
				panic("deprecated. uses 'convert_sections_to_metas_ctx'")
			},
			"convert_sections_to_metas_ctx": func(this *Context, secs []*Section) []*Meta {
				return this.convertSectionToMetas(secs)
			},
			"convert_sections_to_metas_table": func(this *Context, secs []*Section) *MetasTable {
				return this.convertSectionToMetasTable(this.Resource, secs)
			},
			"has_create_permission": hasPermission(this, func(this *Context, p Permissioner) bool {
				return this.hasCreatePermission(p)
			}),
			"has_read_permission": hasPermission(this, func(this *Context, p Permissioner) bool {
				return this.hasReadPermission(p)
			}),
			"has_rec_read_permission": hasRecordPermission(this, func(this *Context) bool {
				return this.HasPermission(this.Resource, roles.Read) && !this.Resource.HasRecordPermission(roles.Read, this.Context, this.Result).Deny()
			}),
			"has_update_permission": hasPermission(this, func(this *Context, p Permissioner) bool {
				return this.hasUpdatePermission(p)
			}),
			"has_rec_update_permission": hasRecordPermission(this, func(this *Context) bool {
				return this.HasPermission(this.Resource, roles.Update) && !this.Resource.HasRecordPermission(roles.Update, this.Context, this.Result).Deny()
			}),
			"has_delete_permission": hasPermission(this, func(this *Context, p Permissioner) bool {
				return this.hasDeletePermission(p)
			}),
			"has_rec_delete_permission": hasRecordPermission(this, func(this *Context) bool {
				return this.HasPermission(this.Resource, roles.Delete) && !this.Resource.HasRecordPermission(roles.Delete, this.Context, this.Result).Deny()
			}),
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
				if this.Type.Has(PRINT) {
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
			"time_loc": this.TimeLocation,
			"crumbs": func() []core.Breadcrumb {
				return this.Breadcrumbs().ItemsWithoutLast()
			},
			"current_crumb": func() core.Breadcrumb {
				return this.Breadcrumbs().Last()
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
				if renderer := GetFrameRenderer(res, name); renderer != nil {
					if err := renderer.Render(this, s); err != nil {
						panic(errors.Wrapf(err, "record frame renderer of %q", name))
					}
					return
				}
				templateNames := GetFrameRendererTemplateName(res, this, name)
				if this.Anonymous() {
					var newTemplateNames []string
					for _, templateName := range templateNames {
						newTemplateNames = append(newTemplateNames, path.Join(path.Dir(templateName), AnonymousDirName, path.Base(templateName)), templateName)
					}
					templateNames = newTemplateNames
				}

				var gt = func(name string) (*template.Executor, error) {
					return this.GetTemplate(name)
				}

				if this.Type.Has(PRINT) {
					gt = func(name string) (*template.Executor, error) {
						return this.GetTemplate(name+".print", name)
					}
				}

				for _, templateName := range templateNames {
					if executor, err := gt(templateName); err == nil {
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

			"action_frame": func(s *template.State, arg *ActionArgument, name string) {
				if renderer := GetFrameRenderer(arg.Action, name); renderer != nil {
					if err := renderer.Render(this, s); err != nil {
						panic(errors.Wrapf(err, "action argument frame renderer of %q", name))
					}
					return
				}
				templateNames := GetFrameRendererTemplateName(arg.Action, this, name)
				if this.Anonymous() {
					var newTemplateNames []string
					for _, templateName := range templateNames {
						newTemplateNames = append(newTemplateNames, path.Join(path.Dir(templateName), AnonymousDirName, path.Base(templateName)), templateName)
					}
					templateNames = newTemplateNames
				}

				var gt = func(name string) (*template.Executor, error) {
					return this.GetTemplate(name)
				}

				if this.Type.Has(PRINT) {
					gt = func(name string) (*template.Executor, error) {
						return this.GetTemplate(name+".print", name)
					}
				}

				for _, templateName := range templateNames {
					if executor, err := gt(templateName); err == nil {
						defer this.WithResult(arg)()
						if err = executor.Execute(s.Writer(), this); err != nil {
							panic(errors.Wrapf(err, "action argument frame render of %q", name))
						}
						return
					} else if !oscommon.IsNotFound(err) {
						panic(errors.Wrapf(err, "action argument frame render of %q", name))
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

			"auth_alternated": func() bool {
				return auth.IsAlternated(this.Admin.Auth.Auth(), this.Context)
			},

			"admin_ctx_set_section_layout": func(ctx *Context, layout string) {
				ctx.SectionLayout = layout
			},

			"admin_ctx_set_type": func(ctx *Context, typ ...interface{}) string {
				ctx.Type = 0
				for _, t := range typ {
					switch t := t.(type) {
					case ContextType:
						ctx.Type |= t
					case string:
						ctx.Type.ParseMerge(t)
					}
				}
				return ""
			},

			"now": func(layout ...string) template.HTML {
				n := this.Now()
				for _, l := range layout {
					return template.HTML(n.Format(ParseTimeLayout(l)))
				}
				return template.HTML(n.String())
			},

			"slice_value_get_deleted_map": func(v interface{}) map[string]bool {
				switch t := v.(type) {
				case *resource.SliceValue:
					return t.DeletedMap()
				}
				return make(map[string]bool, 0)
			},
		},
		this.Admin.funcMaps,
	}

	funcMaps = append(funcMaps, this.funcMaps...)

	return funcMaps
}
