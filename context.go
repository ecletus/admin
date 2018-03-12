package admin

import (
	"bytes"
	"fmt"
	"strings"
	"net/http"
	"path/filepath"
	"errors"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
	"github.com/qor/session"
	"github.com/moisespsena/template/html/template"
	"github.com/moisespsena/template/cache"
)

// Context admin context, which is used for admin controller
type Context struct {
	*qor.Context
	*Searcher
	Resource     *Resource
	Admin        *Admin
	Content      template.HTML
	Action       string
	Settings     map[string]interface{}
	Result       interface{}
	RouteHandler *routeHandler
	PageTitle    string

	usedThemes []string
	funcMaps   []template.FuncMap
	funcValues *template.FuncValues
}

// NewContext new admin context
func (admin *Admin) NewContext(args... interface{}) (c *Context) {
	if len(args) == 0 {
		return admin.NewContext(&qor.Context{})
	}
	for i, arg := range args {
		switch ctx := arg.(type) {
		case qor.SiteInterface:
			return admin.NewContext(ctx.NewContext())
		case *qor.Context:
			c = &Context{Context: ctx, Admin: admin, Settings: map[string]interface{}{}}
		case http.ResponseWriter:
			_, qorCtx := qor.NewContextFromRequestPair(ctx, args[i+1].(*http.Request), admin.router.Prefix)
			qorCtx.Config = admin.Config.Config
			c = &Context{Context: qorCtx, Admin: admin, Settings: map[string]interface{}{}}
		}
	}

	if c != nil {
		if c.Context == nil {
			_, c.Context = qor.NewContextFromRequestPair(c.Writer, c.Request, admin.router.Prefix)
			c.Request = c.Context.Request
		}
		c.PageTitle = admin.SiteTitle

		for _, cb := range admin.NewContextCallbacks {
			c = cb(c)
		}
	}

	return
}

func (admin *Admin) NewContextForResource(context *Context, resource *Resource) *Context {
	clone := admin.NewContext(context.Writer, context.Request)
	clone.Searcher = &Searcher{Context: clone}
	clone.Resource = resource
	clone.DB = clone.DB.NewScope(resource.Value).DB()
	return clone
}

// Funcs set FuncMap for templates
func (context *Context) Funcs(funcMaps... template.FuncMap) *Context {
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
		Context:  context.Context,
		Searcher: context.Searcher,
		Resource: context.Resource,
		Admin:    context.Admin,
		Result:   context.Result,
		Content:  context.Content,
		Settings: context.Settings,
		Action:   context.Action,
		funcMaps: context.funcMaps,
	}
}

func (context *Context) IsAction(name string, names... string) bool {
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

func (context *Context) setResource(res *Resource) *Context {
	if res != nil {
		context.Resource = res
		context.ResourceID = res.GetPrimaryValue(context.Request)
	}
	context.Searcher = &Searcher{Context: context}
	return context
}

func (context *Context) SetResource(res *Resource) *Context {
	return context.setResource(res)
}

func (context *Context) SetResourceWithDB(res *Resource) *Context {
	ctx := context.setResource(res)
	ctx.DB = ctx.DB.NewScope(res.Value).DB()
	return ctx
}

func (context *Context) Asset(layouts ...string) ([]byte, error) {
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
			if content, err := context.Admin.AssetFS.Asset(filepath.Join(prefix, layout)); err == nil {
				return content, nil
			}
		}

		if content, err := context.Admin.AssetFS.Asset(layout); err == nil {
			return content, nil
		}
	}

	return []byte(""), fmt.Errorf("template not found: %v", layouts)
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
	if content, err := context.Asset(name + ".tmpl"); err == nil {
		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, err
		}
		return tmpl.CreateExecutor(), nil
	} else {
		return nil, nil
	}
}

func (context *Context) GetTemplateOrDefault(name string, defaul *template.Executor, others... string) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreNames(name, context.LoadTemplate, others...)
	if t == nil && err == nil {
		return defaul.FuncsValues(context.FuncValues()), nil
	}

	return
}

// renderWith render template based on data
func (context *Context) GetTemplate(name string, others... string) (t *template.Executor, err error) {
	t, err = cache.Cache.LoadOrStoreNames(name, context.LoadTemplate, others...)
	if t == nil && err == nil {
		var msg string
		if len(others) > 0 {
			msg = "Templates with \"" + strings.Join(append([]string{name}, others...), "\", \"") + "\" does not exists."
		} else{
			msg = "Template \"" + name + "\" not exists."
		}
		return nil,  errors.New(msg)
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
		return template.HTML(err.Error())
	}
	return template.HTML(text)
}

// Render render template based on context
func (context *Context) Render(name string, results ...interface{}) template.HTML {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Get error when render file %v: %v", name, r)
			utils.ExitWithMsg(err)
		}
	}()

	clone := context.clone()
	if len(results) > 0 {
		clone.Result = results[0]
	}

	return clone.renderWith(name, clone)
}

// Execute execute template with layout
func (context *Context) Execute(name string, result interface{}) {
	if name == "show" && !context.Resource.isSetShowAttrs {
		name = "edit"
	}

	if context.Action == "" {
		context.Action = name
	}

	var (
		executor *template.Executor
		err error
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
func (context *Context) JSON(action string, result interface{}) {
	if context.Encode(action, result) == nil {
		context.Writer.Header().Set("Content-Type", "application/json")
	}
}

func (context *Context) Encode(action string, result interface{}) error {
	if action == "show" && !context.Resource.isSetShowAttrs {
		action = "edit"
	}

	encoder := Encoder{
		Action:   action,
		Resource: context.Resource,
		Context:  context,
		Result:   result,
	}

	if layout, ok := context.Request.URL.Query()["display"]; ok {
		encoder.Layout = "display." + layout[0]
	}

	return context.Admin.Encode(context.Writer, encoder)
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
	key := "qor_admin.action." + context.Action

	switch context.Action {
	case "new":
		defaul = "Add {{.}}"
	case "edit":
		defaul = "Edit {{.}}"
	case "show":
		defaul = "{{.}} Details"
	default:
		return ""
	}

	return string(context.t(key, defaul))
}