package admin

import (
	"net/http"

	"github.com/aghape/core"
	"github.com/aghape/roles"
	"github.com/moisespsena/go-route"
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
	Permissioner   HasPermissioner
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
	Handle             Handler
	ParentInterseptors [][]func(chain *Chain)
	Interseptors       []func(chain *Chain)
	Config             *RouteConfig
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
func (h *RouteHandler) ServeHTTPContext(w http.ResponseWriter, r *http.Request, rctx *route.RouteContext) {
	context := ContextFromRouteContext(rctx)
	var interseptors []func(chain *Chain)
	for _, p := range h.ParentInterseptors {
		interseptors = append(interseptors, p...)
	}
	interseptors = append(interseptors, h.Interseptors...)
	NewChain(context, h.Handle, interseptors).Next()
}

func (h *RouteHandler) Intercept(f ...func(chain *Chain)) {
	h.Interseptors = append(h.Interseptors, f...)
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
	}

	return &handler
}

func (handler RouteHandler) HasPermission(permissionMode roles.PermissionMode, context *core.Context) bool {
	if handler.Config.Available != nil && !handler.Config.Available(context) {
		return false
	}

	if handler.Config.Permissioner == nil {
		return true
	}

	if handler.Config.PermissionMode != "" {
		return handler.Config.Permissioner.HasPermission(handler.Config.PermissionMode, context)
	}

	return handler.Config.Permissioner.HasPermission(permissionMode, context)
}
