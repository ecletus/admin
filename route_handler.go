package admin

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"

	"github.com/ecletus/roles"

	"github.com/moisespsena-go/xroute"
)

var blankPermissionMode roles.PermissionMode

type DataStack struct {
	Parent *DataStack
	Map    map[interface{}]interface{}
}

func (d *DataStack) ConfigSet(key, value interface{}) {
	if d.Map == nil {
		d.Map = make(map[interface{}]interface{})
	}
	d.Map[key] = value
}

func (d *DataStack) ConfigGet(key interface{}) (value interface{}, ok bool) {
	for d != nil {
		if value, ok = d.Map[key]; ok {
			return
		}
		d = d.Parent
	}
	return
}

func (d *DataStack) NewChild() DataStack {
	return DataStack{d, map[interface{}]interface{}{}}
}

func NoCrumbers(*RouteHandler, *Context, ...string) {}

// RouteConfig config for admin routes
type RouteConfig struct {
	Name     string
	Resource *Resource
	IDParser func(ctx *Context, res *Resource, value string) (ID aorm.ID, err error)
	Permissioner,
	ContextPermissioner core.Permissioner
	PermissionMode                roles.PermissionMode
	Values                        map[interface{}]interface{}
	Data                          DataStack
	Available                     func(context *Context) bool
	CrumbsLoaderFunc              func(rh *RouteHandler, ctx *Context, pattern ...string)
	GlobalPermissionCheckDisabled bool
	HandlerConfig                 interface{}
}

func NewRouteConfig(cfg *RouteConfig) *RouteConfig {
	return cfg
}

func (c *RouteConfig) ConfigSet(key, value interface{}) {
	c.Data.ConfigSet(key, value)
}

func (c *RouteConfig) ConfigGet(key interface{}) (value interface{}, ok bool) {
	return c.Data.ConfigGet(key)
}

func (c *RouteConfig) Options(opt ...core.Option) *RouteConfig {
	for _, opt := range opt {
		opt.Apply(c)
	}
	return c
}

func (c *RouteConfig) LoadCrumbs(rh *RouteHandler, ctx *Context, pattern ...string) {
	if c.CrumbsLoaderFunc != nil {
		c.CrumbsLoaderFunc(rh, ctx, pattern...)
	}
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

func (h *RouteHandler) ServeHTTPContext(w http.ResponseWriter, _ *http.Request, rctx *xroute.RouteContext) {
	var (
		err            error
		context        = ContextFromRouteContext(rctx)
		currentUser, _ = context.Admin.Auth.GetCurrentUser(context)
	)

	if currentUser == nil && !h.Config.GlobalPermissionCheckDisabled {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	context.SetCurrentUser(currentUser)
	if h.Config.Resource != nil {
		if h.Config.IDParser != nil {
			context.IDParser = h.Config.IDParser
		}
		context.SetResource(h.Config.Resource)
	}

	context.Role.Register(roles.GetVisitor())

	if currentUser != nil {
		context.DB(context.DB().SetCurrentUser(aorm.IDOf(currentUser)))
		var superUser bool
		if superUser, err = context.Admin.Auth.IsSuperAdmin(context); err != nil {
			http.Error(context.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if context.Roles.Len() == 0 {
			// all user roles
			context.Roles = context.Role.MatchedRoles(context.Request, currentUser)

			if superUser {
				// all roles
				context.Roles.Append(roles.Anyone)
			}
		} else {
			// only roles with exists for user
			context.Roles = context.Role.MatchedRoles(context.Request, currentUser).Intersection(context.Roles.Strings())
		}
	} else {
		context.Roles = context.Role.MatchedRoles(context.Request, nil)
	}

	if setuper, ok := context.Admin.Auth.(ContextSetuper); ok {
		if err = setuper.ContextSetup(context); err != nil {
			panic(err)
		}
	}

	if !h.Config.GlobalPermissionCheckDisabled && !context.HasPermission(core.Permissioners(h, context.Admin.Auth), context.PermissionMode) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	context.RouteHandler = h
	h.generateCrumbs(context)

	if context.HasError() {
		http.Error(w, context.Error(), http.StatusInternalServerError)
		return
	}

	for _, res := range context.ParentResource {
		if res.ContextSetuper != nil {
			if err = res.ContextSetuper.ContextSetup(context); err != nil {
				panic(errors.Wrapf(err, "%q Resource context setup", res.UID))
			}
		}
	}

	if perm := h.HasContextPermission(context.PermissionMode, context.Context); perm.Deny() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if perm := context.Admin.HasContextPermission(context.PermissionMode, context); perm.Deny() {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

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

func (h RouteHandler) HasPermission(mode roles.PermissionMode, ctx *core.Context) (perm roles.Perm) {
	context := ContextFromCoreContext(ctx)
	if h.Config.Available != nil && !h.Config.Available(context) {
		return roles.DENY
	}

	if h.Config.Permissioner == nil {
		return
	}

	if h.Config.PermissionMode != "" {
		mode = h.Config.PermissionMode
	}

	return h.Config.Permissioner.HasPermission(mode, context.Context)
}

func (h *RouteHandler) HasContextPermission(mode roles.PermissionMode, ctx *core.Context) (perm roles.Perm) {
	if h.Config.ContextPermissioner == nil {
		return
	}
	context := ContextFromCoreContext(ctx)

	if h.Config.PermissionMode != "" {
		mode = h.Config.PermissionMode
	}
	return h.Config.ContextPermissioner.HasPermission(mode, context.Context)
}

func NewHandler(handle Handler, configs ...*RouteConfig) *RouteHandler {
	handler := &RouteHandler{
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
		if handler.Config.CrumbsLoaderFunc == nil {
			handler.CrumbsLoader = handler.Config.Resource.GetAdmin()
		}
	}
	if handler.CrumbsLoader == nil {
		handler.CrumbsLoader = handler.Config
	}
	return handler
}
