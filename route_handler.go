package admin

import (
	"net/http"

	"github.com/ecletus/core"
	"github.com/ecletus/roles"
	"github.com/moisespsena-go/xroute"
)

var blankPermissionMode roles.PermissionMode

type DataStack struct {
	Parent *DataStack
	Map    map[interface{}]interface{}
}

func (d *DataStack) NewChild() DataStack {
	return DataStack{d, map[interface{}]interface{}{}}
}

// RouteConfig config for admin routes
type RouteConfig struct {
	Resource       *Resource
	Permissioner   core.Permissioner
	PermissionMode roles.PermissionMode
	Values         map[interface{}]interface{}
	Data           DataStack
	Available      func(context *core.Context) bool
}

func (c *RouteConfig) Set(key, value interface{}) {
	if c.Data.Map == nil {
		c.Data.Map = map[interface{}]interface{}{}
	}
	c.Data.Map[key] = value
}

func (c *RouteConfig) Get(key interface{}) interface{} {
	d := &c.Data
	for d != nil {
		if v, ok := d.Map[key]; ok {
			return v
		}
		d = d.Parent
	}
	return nil
}

type Handler func(c *Context)

type Chain struct {
	Context  *Context
	Handler  Handler
	Handlers []func(chain *Chain)
	index    int
	pass     bool
}

func NewChain(context *Context, handler Handler, handlers []func(chain *Chain)) *Chain {
	return &Chain{Context: context, Handler: handler, Handlers: handlers}
}

func (c *Chain) Pass() {
	c.pass = true
}

func (c *Chain) Next() {
	old := c.Context
	defer func() {
		c.Context = old
	}()

	l := len(c.Handlers)
	for c.index < l {
		h := c.Handlers[c.index]
		c.index++
		h(c)
		if c.pass {
			c.pass = false
		} else {
			return
		}
	}
	if c.index == l {
		c.index++
		c.Handler(c.Context)
	}
}

type RouteHandler struct {
	Path               string
	Handle             Handler
	ParentInterseptors [][]func(chain *Chain)
	Interseptors       []func(chain *Chain)
	Config             *RouteConfig
	Name               string
	CrumbsLoader       CrumbsLoader
}

func (h *RouteHandler) SetName(name string) *RouteHandler {
	h.Name = name
	return h
}

func (h RouteHandler) WithName(name string) *RouteHandler {
	h.Name = name
	return &h
}

func (h RouteHandler) Clone() *RouteHandler {
	h.Interseptors = h.Interseptors[:]
	h.Config = &(*h.Config)
	h.Config.Data = h.Config.Data.NewChild()
	return &h
}

func (h RouteHandler) Child() *RouteHandler {
	h.ParentInterseptors = append(h.ParentInterseptors, h.Interseptors)
	h.Interseptors = nil
	h.Config = &(*h.Config)
	h.Config.Data = h.Config.Data.NewChild()
	return &h
}
func (h *RouteHandler) ServeHTTPContext(w http.ResponseWriter, r *http.Request, rctx *xroute.RouteContext) {
	context := ContextFromRouteContext(rctx)
	if !core.HasPermission(h, context.PermissionMode, context.Context) {
		context.Writer.WriteHeader(http.StatusForbidden)
		context.Writer.Write([]byte(`Forbidden`))
		return
	}
	context.RouteHandler = h
	var interseptors []func(chain *Chain)
	for _, p := range h.ParentInterseptors {
		interseptors = append(interseptors, p...)
	}
	interseptors = append(interseptors, h.Interseptors...)
	h.generateCrumbs(context)
	NewChain(context, h.Handle, interseptors).Next()
}

func (h *RouteHandler) Intercept(f ...func(chain *Chain)) {
	h.Interseptors = append(h.Interseptors, f...)
}

func (h RouteHandler) HasPermissionE(permissionMode roles.PermissionMode, context *core.Context) (ok bool, err error) {
	if h.Config.Available != nil && !h.Config.Available(context) {
		return false, nil
	}

	if h.Config.Permissioner == nil {
		return true, nil
	}

	if h.Config.PermissionMode != "" {
		permissionMode = h.Config.PermissionMode
	}

	return h.Config.Permissioner.HasPermissionE(permissionMode, context)
}

func NewHandler(handle Handler, configs ...*RouteConfig) *RouteHandler {
	handler := RouteHandler{
		Handle: handle,
	}

	for _, config := range configs {
		handler.Config = config
	}

	if handler.Config == nil {
		handler.Config = &RouteConfig{}
	}

	if handler.Config.Permissioner == nil && handler.Config.Resource != nil {
		handler.Config.Permissioner = handler.Config.Resource
	}

	if handler.Config.Resource != nil {
		handler.Config.Resource.mounted = true
		handler.CrumbsLoader = handler.Config.Resource.GetAdmin()
	}

	return &handler
}
