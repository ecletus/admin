package admin

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/moisespsena/go-assetfs"
	"github.com/moisespsena/go-assetfs/api"
	"github.com/moisespsena/go-route"
	"github.com/moisespsena/template/cache"
	"github.com/moisespsena/template/html/template"
	"github.com/aghape/core"
	"github.com/aghape/core/utils"
	"github.com/aghape/responder"
	"github.com/aghape/roles"
	"github.com/aghape/session"
)

type ContextType string

func (ct ContextType) String() string {
	return string(ct)
}

func (ct ContextType) S() string {
	return string(ct)
}

const (
	NONE  ContextType = ""
	INDEX ContextType = "index"
	NEW   ContextType = "new"
	SHOW  ContextType = "show"
	EDIT  ContextType = "edit"
)

// Context admin context, which is used for admin controller
type Context struct {
	*core.Context
	*Searcher
	Scheme       *Scheme
	Resource     *Resource
	ResourceType string
	Admin        *Admin
	Content      template.HTML
	Action       string
	Settings     map[string]interface{}
	Result       interface{}
	PageTitle    string

	usedThemes     []string
	funcMaps       []template.FuncMap
	funcValues     *template.FuncValues
	PermissionMode roles.PermissionMode
	Display        string
	Type           ContextType
}

const (
	P_LAYOUT  = "_layout"
	P_DISPLAY = "_display"
)

// NewContext new admin context
func (admin *Admin) NewContext(args ...interface{}) (c *Context) {
	if len(args) == 0 {
		return admin.NewContext(&core.Context{})
	}
	for i, arg := range args {
		switch ctx := arg.(type) {
		case core.SiteInterface:
			return admin.NewContext(ctx.NewContext())
		case *core.Context:
			c = &Context{Context: ctx}
		case http.ResponseWriter:
			_, qorCtx := core.NewContextFromRequestPair(ctx, args[i+1].(*http.Request), admin.Router.Prefix())
			qorCtx.Config = admin.Config.Config
			c = &Context{Context: qorCtx}
		}
	}

	if c != nil {
		if c.Context == nil {
			_, c.Context = core.NewContextFromRequestPair(c.Writer, c.Request, admin.Router.Prefix())
			c.Request = c.Context.Request
		}
		c.Settings = map[string]interface{}{}
		c.Admin = admin
		c.Context.Data().Set(CONTEXT_KEY, c)
		//c.PageTitle = admin.SiteTitle

		if c.Request != nil {
			if v := c.Request.URL.Query().Get(P_DISPLAY); v != "" {
				c.Display = v
			}
		}

		for _, cb := range admin.NewContextCallbacks {
			c = cb(c)
		}
	}

	return
}

func (admin *Admin) NewContextForResource(context *Context, resource *Resource) *Context {
	clone := &(*context)
	clone.Searcher = clone.NewSearcher()
	clone.Resource = resource
	clone.Scheme = resource.Scheme
	clone.DB = clone.DB.NewScope(resource.Value).DB()
	return clone
}

func (context *Context) WithResource(res *Resource, value interface{}) func() {
	id, parentResourceID, resource, searcher, DB, result, scheme := context.ResourceID,
		context.ParentResourceID, context.Resource, context.Searcher, context.DB, context.Result, context.Scheme

	newDB := DB
	if context.Context.Parent != nil {
		newDB = context.Context.Parent.DB
	}

	context.ResourceID,
		context.ParentResourceID, context.Resource, context.Searcher, context.DB, context.Result, context.Scheme = "",
		[]string{}, res, context.NewSearcher(), newDB, value, res.Scheme
	if value != nil {
		context.ResourceID = res.GetKey(value)
	}
	return func() {
		context.ResourceID,
			context.ParentResourceID, context.Resource, context.Searcher, context.DB, context.Result, context.Scheme = id,
			parentResourceID, resource, searcher, DB, result, scheme
	}
}

func (context *Context) HtmlifyRecordsMeta(res *Resource, metaName string, records ...interface{}) (result []template.HTML) {
	if len(records) == 0 {
		return
	}
	defer context.WithResource(res, nil)()
	result = make([]template.HTML, len(records))
	qorContext := context.CloneBasic()
	valuer := res.GetDefinedMeta(metaName).GetFormattedValuer()
	var v interface{}
	for i, item := range records {
		if item == nil {
			continue
		}
		context.ResourceID = res.GetKey(item)
		v = valuer(item, qorContext)
		if v != nil {
			switch vt := v.(type) {
			case template.HTML:
				result[i] = vt
			case string:
				result[i] = template.HTML(vt)
			default:
				result[i] = context.HtmlifyInterfaces(v)[0]
			}
		}
	}
	return
}

func (context *Context) HtmlifyRecordMeta(res *Resource, metaName string, record interface{}) template.HTML {
	return context.HtmlifyRecordsMeta(res, metaName, record)[0]
}

func (context *Context) HtmlifyRecords(res *Resource, records ...interface{}) (result []template.HTML) {
	return context.HtmlifyRecordsMeta(res, BASIC_META_HTML, records...)
}

func (context *Context) HtmlifyRecord(res *Resource, record interface{}) template.HTML {
	return context.HtmlifyRecordMeta(res, BASIC_META_HTML, record)
}

func (context *Context) HtmlifyInterfaces(values ...interface{}) (result []template.HTML) {
	result = make([]template.HTML, len(values))
	for i, value := range values {
		if value == nil {
			continue
		}
		switch vt := value.(type) {
		case interface{ Htmlify(*Context) template.HTML }:
			result[i] = vt.Htmlify(context)
		default:
			result[i] = context.Context.Htmlify(value)
		}
	}
	return
}

func (context *Context) HtmlifyInterface(value interface{}) (result template.HTML) {
	if value == nil {
		return
	}
	switch vt := value.(type) {
	case interface{ Htmlify(*Context) template.HTML }:
		return vt.Htmlify(context)
	default:
		return context.Context.Htmlify(value)
	}
}

func (context *Context) HtmlifyItems(values ...interface{}) (result []template.HTML) {
	if l := len(values); l > 0 {
		if res, ok := values[0].(*Resource); ok {
			if l == 1 {
				return
			}
			return context.HtmlifyRecords(res, values[1:]...)
		}
	}
	return context.HtmlifyInterfaces(values...)
}

func (context *Context) Htmlify(value interface{}, res ...*Resource) (result template.HTML) {
	if len(res) > 0 {
		return context.HtmlifyRecords(res[0], value)[0]
	}
	return context.HtmlifyInterfaces(value)[0]
}

func (context *Context) ValidateLayout() bool {
	l := context.Resource.GetLayout(context.Layout)
	if l == nil {
		context.AddError(fmt.Errorf("Layout %q is not valid.", context.Layout))
		return false
	}
	return true
}

func (context *Context) ValidateLayoutOrError() bool {
	if !context.ValidateLayout() {
		context.SendError()
		return false
	}
	return true
}

func (context *Context) LoadDisplay(displayType string) bool {
	if context.HasError() {
		return false
	}

	if context.Display != "" {
		display := context.Resource.GetDisplay(displayType + "/" + context.Display)
		if display == nil {
			context.AddError(fmt.Errorf("Display %q does not exists.", context.Display))
		} else {
			context.Layout = display.GetLayoutName()
		}
		context.ValidateLayout()
	} else {
		context.Display = context.Resource.GetDefaultDisplayName()
	}
	return !context.HasError()
}

func (context *Context) TypeS() string {
	return context.Type.S()
}

func (context *Context) Is(values ...interface{}) bool {
	for _, v := range values {
		switch vt := v.(type) {
		case ContextType:
			if context.Type == vt {
				return true
			}
		case string:
			if context.Type.S() == vt {
				return true
			}
		}
	}
	return false
}

func (context *Context) LoadDisplayOrError(displayType ...string) bool {
	if len(displayType) == 0 || displayType[0] == "" {
		displayType = []string{context.Type.S()}
	}
	if !context.LoadDisplay(displayType[0]) {
		context.Writer.WriteHeader(http.StatusPreconditionFailed)
		return false
	}
	return true
}

func (context *Context) CreateChild(res *Resource, record ...interface{}) *Context {
	context = context.clone()
	_, context.Context = context.Context.NewChild(nil)
	context.Resource = res
	context.ParentResourceID = []string{}
	if len(record) == 1 && record[0] != nil {
		context.Result = record[0]
		context.ResourceID = res.GetKey(record[0])
	} else {
		context.Result = nil
		context.ResourceID = ""
	}
	return context
}

// Funcs set FuncMap for templates
func (context *Context) Funcs(funcMaps ...template.FuncMap) *Context {
	context.funcMaps = append(context.funcMaps, funcMaps...)
	return context
}

// Flash set flash message
func (context *Context) Flash(message string, typ string) {
	context.SessionManager().Flash(session.Message{
		Message: template.HTML(message),
		Type:    typ,
	})
}

func (context *Context) clone() *Context {
	return &Context{
		Context:      context.Context,
		Searcher:     context.Searcher,
		Resource:     context.Resource,
		ResourceType: context.ResourceType,
		Admin:        context.Admin,
		Result:       context.Result,
		Content:      context.Content,
		Settings:     context.Settings,
		Action:       context.Action,
		funcMaps:     context.funcMaps,
		PageTitle:    context.PageTitle,
		Type:         context.Type,
	}
}

func (context *Context) IsAction(name string, names ...string) bool {
	if context.Action == name {
		return true
	}

	for _, name = range names {
		if context.Action == name {
			return true
		}
	}

	return false
}

// Get get context's Settings
func (context *Context) Get(key string) interface{} {
	return context.Settings[key]
}

// Set set context's Settings
func (context *Context) Set(key string, value interface{}) {
	context.Settings[key] = value
}

func (context *Context) resourcePath() string {
	if context.Resource == nil {
		return ""
	}
	return context.Resource.ToParam()
}

func (context *Context) NewSearcher() *Searcher {
	s := &Searcher{Context: context}
	if context.Request != nil {
		if layout, ok := context.Request.URL.Query()[P_LAYOUT]; ok {
			s.Layout = layout[len(layout)-1]
		}
	}
	return s
}

func (context *Context) setResource(res *Resource, recorde ...interface{}) *Context {
	if res != nil {
		context.Resource = res
		context.Scheme = res.Scheme
		if len(recorde) == 1 && recorde[1] != nil {
			context.ResourceID = res.GetKey(recorde)
		} else {
			context.ResourceID = context.URLParam(res.ParamIDName())
		}
	}
	context.Searcher = context.NewSearcher()
	return context
}

func (context *Context) SetResource(res *Resource, recorde ...interface{}) *Context {
	return context.setResource(res)
}

func (context *Context) SetResourceWithDB(res *Resource) *Context {
	ctx := context.setResource(res)
	ctx.DB = ctx.DB.NewScope(res.Value).DB()
	return ctx
}

func (context *Context) Asset(layouts ...string) (asset assetfs.AssetInterface, err error) {
	return context.getAsset(context.Admin.TemplateFS, layouts...)
}

func (context *Context) StaticAsset(layouts ...string) (asset assetfs.AssetInterface, err error) {
	return context.getAsset(context.Admin.StaticFS, layouts...)
}

func (context *Context) getAsset(fs assetfs.Interface, layouts ...string) (asset assetfs.AssetInterface, err error) {
	var prefixes, themes []string

	if context.Request != nil {
		if theme := context.Request.URL.Query().Get("theme"); theme != "" {
			themes = append(themes, theme)
		}
	}

	if len(themes) == 0 && context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			themes = append(themes, theme.GetName())
		}
	}

	if resourcePath := context.resourcePath(); resourcePath != "" {
		for _, theme := range themes {
			prefixes = append(prefixes, filepath.Join("themes", theme, resourcePath))
		}
		prefixes = append(prefixes, resourcePath)
	}

	for _, theme := range themes {
		prefixes = append(prefixes, filepath.Join("themes", theme))
	}

	for _, layout := range layouts {
		for _, prefix := range prefixes {
			if asset, err = fs.Asset(filepath.Join(prefix, layout)); err == nil {
				return
			}
		}

		if asset, err = fs.Asset(layout); err == nil {
			return
		}
	}

	return nil, fmt.Errorf("template not found: %v", layouts)
}

// renderText render text based on data
func (context *Context) renderText(text string, data interface{}) template.HTML {
	var (
		err    error
		tmpl   *template.Template
		result = bytes.NewBufferString("")
	)

	if tmpl, err = template.New("").Parse(text); err == nil {
		if err = context.ExecuteTemplate(tmpl, result, data); err == nil {
			return template.HTML(result.String())
		}
	}

	return template.HTML(err.Error())
}

func (context *Context) LoadTemplate(name string) (*template.Executor, error) {
	if asset, err := context.Asset(name + ".tmpl"); err == nil {
		tmpl, err := template.New(name).SetPath(asset.GetPath()).Parse(asset.GetString())
		if err != nil {
			return nil, err
		}
		return tmpl.CreateExecutor(), nil
	} else {
		return nil, nil
	}
}

func (context *Context) LoadTemplateInfo(info api.FileInfo) (*template.Executor, error) {
	data, err := info.Data()
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New(info.Name()).SetPath(info.RealPath()).Parse(string(data))
	if err != nil {
		return nil, err
	}
	return tmpl.CreateExecutor(), nil
}

func (context *Context) GetTemplateOrDefault(name string, defaul *template.Executor, others ...string) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreNames(name, context.LoadTemplate, others...)
	if t == nil && err == nil {
		return defaul.FuncsValues(context.FuncValues()), nil
	}

	return
}

// renderWith render template based on data
func (context *Context) GetTemplate(name string, others ...string) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreNames(name, context.LoadTemplate, others...)
	if t == nil && err == nil {
		var msg string
		if len(others) > 0 {
			msg = "Templates with \"" + strings.Join(append([]string{name}, others...), "\", \"") + "\" does not exists."
		} else {
			msg = "Template \"" + name + "\" not exists."
		}
		return nil, errors.New(msg)
	}

	return t.FuncsValues(context.FuncValues()), nil
}

// renderWith render template based on data
func (context *Context) GetTemplateInfo(info api.FileInfo, others ...api.FileInfo) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreInfos(info, context.LoadTemplateInfo, others...)
	if err != nil {
		return nil, err
	}
	return t.FuncsValues(context.FuncValues()), nil
}

// renderWith render template based on data
func (context *Context) renderWith(name string, data interface{}) template.HTML {
	executor, err := context.GetTemplate(name)
	if err != nil {
		return template.HTML(err.Error())
	}

	text, err := executor.ExecuteString(data)
	if err != nil {
		if et, ok := err.(*template.ErrorWithTrace); ok {
			panic(et)
		}
		return template.HTML(err.Error())
	}
	return template.HTML(text)
}

// renderWith render template based on data
func (context *Context) renderWithInfo(info api.FileInfo, data interface{}) template.HTML {
	executor, err := context.GetTemplateInfo(info)
	if err != nil {
		return template.HTML(err.Error())
	}
	text, err := executor.ExecuteString(data)
	if err != nil {
		if et, ok := err.(*template.ErrorWithTrace); ok {
			panic(et)
		}
		return template.HTML(err.Error())
	}
	return template.HTML(text)
}

// Render render template based on context
func (context *Context) Render(name string, results ...interface{}) template.HTML {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Get error when render file %v:\n%v", name, r)
			if et, ok := r.(*template.ErrorWithTrace); ok {
				et.Err = err.Error()
				panic(et)
			}
			utils.ExitWithMsg(err)
		}
	}()

	clone := context.clone()
	if len(results) > 0 {
		clone.Result = results[0]
	}

	return clone.renderWith(name, clone)
}

// Render render template based on context
func (context *Context) RenderInfo(info api.FileInfo, results ...interface{}) template.HTML {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Get error when render file %v:\n%v", info.RealPath(), r)
			if et, ok := r.(*template.ErrorWithTrace); ok {
				et.Err = err.Error()
				panic(et)
			}
			utils.ExitWithMsg(err)
		}
	}()

	clone := context.clone()
	if len(results) > 0 {
		clone.Result = results[0]
	}

	return clone.renderWithInfo(info, clone)
}

// Execute execute template with layout
func (context *Context) Execute(name string, result interface{}) {
	if name == "" {
		name = context.Type.S()
	}

	if name == "show" && !context.Resource.isSetShowAttrs {
		name = "edit"
	}

	if context.Action == "" {
		context.Action = name
	}

	var (
		executor *template.Executor
		err      error
	)

	if executor, err = context.GetTemplate("layout"); err != nil {
		utils.ExitWithMsg(err)
	}

	context.Result = result
	context.Content = context.Render(name, result)
	if err := executor.Execute(context.Writer, context); err != nil {
		utils.ExitWithMsg(err)
	}
}

// JSON generate json outputs for action
func (context *Context) JSON(result interface{}, action ...string) {
	if context.Encode(result, action...) == nil {
		context.Writer.Header().Set("Content-Type", "application/json")
	}
}

func (context *Context) Encode(result interface{}, layout ...string) error {
	if len(layout) == 0 {
		layout = []string{context.Layout}
	}
	if layout[0] == "show" && !context.Resource.isSetShowAttrs {
		layout[0] = "edit"
	}

	encoder := Encoder{
		Layout:   layout[0],
		Resource: context.Resource,
		Context:  context,
		Result:   result,
	}

	return context.Admin.Encode(context.Writer, encoder)
}

func (context *Context) SendError() bool {
	if context.HasError() {
		responder.With("html", func() {
			context.Flash(context.Error(), "error")
		}).With("json", func() {
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
		return true
	}
	return false
}

// GetSearchableResources get defined searchable resources has performance
func (context *Context) GetSearchableResources() (resources []*Resource) {
	if admin := context.Admin; admin != nil {
		for _, res := range admin.searchResources {
			if res.HasPermission(roles.Read, context.Context) {
				resources = append(resources, res)
			}
		}
	}
	return
}

// GetSearchableResources clone the context object
func CloneContext(context *Context) *Context {
	return context.clone()
}

func (context *Context) GetActionLabel() string {
	var defaul string
	key := I18NGROUP + ".action." + context.Action

	switch context.Type {
	case NEW:
		defaul = "Add {{.}}"
	case EDIT:
		defaul = "Edit {{.}}"
	case SHOW:
		defaul = "{{.}} Details"
	default:
		return ""
	}

	return string(context.t(key, defaul))
}

func ContextFromQorContext(ctx *core.Context) *Context {
	return ctx.Data().Get(CONTEXT_KEY).(*Context)
}

func ContextFromQorContextOrNew(ctx *core.Context, admin *Admin) *Context {
	c, ok := ctx.Data().GetOk(CONTEXT_KEY)
	if ok {
		return c.(*Context)
	}
	return admin.NewContext(ctx)
}

var CONTEXT_KEY = PKG + ".context"

func ContextFromChain(chain *route.ChainHandler) *Context {
	return ContextFromRouteContext(chain.Context)
}

func SetContexToChain(chain *route.ChainHandler, context *Context) {
	SetContextToRouteContext(chain.Context, context)
}

func ContextFromRouteContext(rctx *route.RouteContext) *Context {
	v, ok := rctx.Data[CONTEXT_KEY]
	if ok {
		return v.(*Context)
	}
	return nil
}

func SetContextToRouteContext(rctx *route.RouteContext, context *Context) {
	rctx.Data[CONTEXT_KEY] = context
}
