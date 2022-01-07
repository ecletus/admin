package admin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/moisespsena/template/html/template"

	"github.com/moisespsena-go/middleware"

	"github.com/ecletus/roles"

	"github.com/ecletus/auth"

	"github.com/moisespsena-go/xroute"

	"github.com/ecletus/core"
)

const (
	AuthRevertPath = "/!auth/revert"
)

type Interseptor func(w http.ResponseWriter, req *http.Request, serv func(w http.ResponseWriter, req *http.Request))

// routeInterseptor dispatches the handler registered in the matched route
func (this *Admin) routeInterseptor(chain *xroute.ChainHandler) {
	mainContext := core.ContextFromRequest(chain.Request())
	req, childContext := mainContext.NewChild(nil, this.Config.MountPath)
	childContext.SetRawDB(childContext.DB().Set(DbKey, this))
	childContext.StaticURL = mainContext.StaticURL + "/admin"
	context := this.NewContext(childContext)
	context.RouteContext = chain.Context
	context.MetaContextFactory = func(parent *core.Context, res interface{}, record interface{}) *core.Context {
		if res, ok := res.(*Resource); ok {
			return ContextFromCoreContext(parent).CreateChild(res, record).Context
		}
		return parent
	}
	childContext.Parent.SetValue(CONTEXT_KEY, context)
	chain.SetRequest(req)

	switch req.Method {
	case http.MethodDelete, http.MethodPost, http.MethodPut:
		// Parse Request Form
		req.ParseMultipartForm(8 * 1024 * 1024)

		// Set Request Method
		if method := req.Form.Get("_method"); method != "" {
			req.Method = strings.ToUpper(method)
			chain.Context.RouteMethod = req.Method
		}
	}

	do := func() {
		context.Breadcrumbs().Append(core.NewBreadcrumb(context.Path(), context.Admin.Config.DefaultPageTitle(context), ""))

		oldKey := chain.Context.DefaultValueKey
		defer func() {
			chain.Context.DefaultValueKey = oldKey
		}()
		chain.Context.DefaultValueKey = CONTEXT_KEY
		chain.Next()
	}

	if this.Config.Public {
		do()
	} else {
		auth.Authenticates(this.Auth, chain.Writer, req, func(ok bool) {
			do()
		})
	}
}

func (this *Admin) handlerInterseptor(chain *xroute.ChainHandler) {
	context := ContextFromCoreContext(core.ContextFromRequest(chain.Request()))
	context.RouteContext = chain.Context

	if context.PermissionMode == roles.NONE {
		switch context.Request.Method {
		case "GET":
			context.PermissionMode = roles.Read
		case "PUT":
			context.PermissionMode = roles.Update
		case "POST":
			context.PermissionMode = roles.Create
		case "DELETE":
			context.PermissionMode = roles.Delete
		}
	}

	chain.Pass()
}

func (this *Admin) InitRoutes(router xroute.Router) {
	router.Intersept(&xroute.Middleware{
		Name:    PKG,
		Handler: this.routeInterseptor,
	})

	adminController := this.Config.Controller
	if adminController == nil {
		adminController = &AdminController{Admin: this}
	} else {
		adminController.Admin = this
	}
	router.Get("/", NewHandler(adminController.Dashboard, &RouteConfig{
		GlobalPermissionCheckDisabled: this.Config.Public,
	}))
	router.Get(AuthRevertPath, NewHandler(func(c *Context) {
		Auth := c.Admin.Auth.Auth()
		if err := auth.Revert(Auth, c.Context); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusPreconditionFailed)
		} else {
			to := c.Request.Header.Get("Referer")
			if to == "" {
				to = c.Path()
			}
			http.Redirect(c.Writer, c.Request, c.Path(), http.StatusSeeOther)
		}
	}))
	router.Get("/search", NewHandler(adminController.SearchCenter))
	router.HandleM(xroute.GET|xroute.HEAD, "/(assets|themes)/", this.StaticFS)

	browserUserAgentRegexp := regexp.MustCompile("Mozilla|Gecko|WebKit|MSIE|Opera")
	router.Use(&xroute.Middleware{
		Name: PKG + ".csrf_check",
		Handler: func(chain *xroute.ChainHandler) {
			request := chain.Request()
			if request.Method != http.MethodGet && request.Method != http.MethodHead {
				if browserUserAgentRegexp.MatchString(request.UserAgent()) {
					if referrer := request.Referer(); referrer != "" {
						if r, err := url.Parse(referrer); err == nil {
							if r.Host == request.Host {
								goto ok
							}
						}
					}
					chain.Writer.Write([]byte("Could not authorize you because 'CSRF detected'"))
					return
				}
			}
		ok:
			chain.Pass()
		},
	})

	router.Use(&xroute.Middleware{
		Name: PKG + ".handler",
		Handler: func(chain *xroute.ChainHandler) {
			middleware.SetNoCache(chain.Writer, chain.Request())
			chain.Pass()
		},
	})

	router.HandlerIntersept(&xroute.Middleware{
		Name:    PKG + ".handler_interseptor",
		Handler: this.handlerInterseptor,
	})

	for _, m := range this.menus {
		if m.Resource == nil && m.URI != "" {
			if m.URI == "/" {
				this.RouteTree.Menu = m
			} else {
				//this.RouteTree.Add(strings.TrimPrefix(m.URI, "/"), &RouteNode{Menu: m})
			}
		}
	}

	var root = &this.RouteTree.RouteNode
	root.Handler = router

	if this.RouteTree.Children == nil {
		this.RouteTree.Children = map[string]*RouteNode{}
	}
	for param, res := range this.ResourcesByParam {
		pattern := "/" + param
		r := res.InitRoutes(root)
		router.Mount(pattern, r)
	}

	for _, r := range this.onRouter {
		r(router)
	}

	var buf bytes.Buffer
	buf.WriteString("Route tree:\n")

	(&RouteNode{Children: map[string]*RouteNode{"": root}}).Walk(func(parents []*RouteNodeWalkerItem, item *RouteNodeWalkerItem) (err error) {
		localPath, label := item.Node.Pair()

		if item.Node.IsIndex() {
			if item.Node.Parent != nil && item.Node.Parent.Handler != nil {
				if mux, ok := item.Node.Parent.Handler.(xroute.Router); ok {
					mux.Get("/"+strings.TrimSuffix(localPath, "/"), NewHandler(func(c *Context) {
						c.Result = item.Node
						adminController.MenuIndex(c, item.Node)
					}))
				}
			}
		}

		var (
			prefix string
			args   = []interface{}{}
		)

		if item.Node.Depth > 0 {
			for _, s := range parents {
				if !s.Last {
					prefix += "│   "
				} else {
					prefix += "   "
				}
			}
			if item.Last {
				prefix += "└──"
			} else {
				prefix += "├──"
			}

			args = append(args, prefix)
		}

		if localPath != "" && label != "" {
			args = append(args, localPath, "→", label)
		} else if label != "" {
			args = append(args, "+", "→", label)
		} else if localPath != "" {
			args = append(args, localPath)
		}
		fmt.Fprintln(&buf, args...)
		return nil
	})

	rlog.Debug(buf.String())

	router.NotFound(NewHandler(this.UserPagesHandler, &RouteConfig{
		GlobalPermissionCheckDisabled: true,
	}))
}

func (this *Admin) UserPagesHandler(ctx *Context) {
	pth := ctx.Request.URL.Path[1:]

	// hidden file
	if pth[0] == '.' || strings.Contains(pth, "/.") {
		return
	}

	var baseDir = "www/"
	if ctx.Anonymous() {
		baseDir += AnonymousDirName + "/"
	}
	pth = path.Join(baseDir, pth)
	if ext := path.Ext(pth); ext == "" {
		pth += ".tmpl"
	}
	storage := ctx.Site.SystemStorage()
	f, err := storage.Get(pth)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		panic(errors.Wrap(err, "get file from system storage"))
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		panic(errors.Wrap(err, "get stat of file from system storage"))
	}
	if strings.HasSuffix(pth, ".tmpl") {
		f.Close()
		var createExecutor func(f *os.File) (*template.Executor, error)
		var paths []string
		var ew = func(w io.Writer, err error, msg string, args ...interface{}) error {
			var m string
			if len(paths) > 0 {
				m = "template {`" + strings.Join(paths, "`/`") + "`}"
			}
			if msg != "" {
				if m != "" {
					m += ": "
				}
				m += fmt.Sprintf(msg, args...)
			}
			if err == nil {
				err = errors.New(m)
			}
			err = errors.Wrap(err, m)
			w.Write([]byte("ERROR: " + err.Error()))
			return err
		}
		include := func(w io.Writer, pth string, dot ...interface{}) {
			if strings.Contains(pth, "..") {
				panic(ew(w, nil, "bad path %q", pth))
			}
			if pth[0] == '/' {
				pth = baseDir + pth[1:]
			} else {
				pth = path.Dir(paths[len(paths)-1]) + "/" + pth
			}
			pth += ".tmpl"
			f, err := storage.Get(pth)
			if err != nil {
				panic(ew(w, err, "get file %q from system storage", pth))
			}
			defer f.Close()
			paths = append(paths, pth)
			defer func() {
				paths = paths[0 : len(paths)-1]
			}()
			exc, err := createExecutor(f)
			if err != nil {
				panic(ew(w, err, "create executor"))
			}
			if len(dot) == 0 {
				dot = append(dot, ctx)
			}
			err = exc.Execute(w, dot[0])
			if err != nil {
				panic(ew(w, err, ""))
			}
		}
		createExecutor = func(f *os.File) (*template.Executor, error) {
			stat, err := f.Stat()
			if err != nil {
				return nil, errors.Wrap(err, "get stat of file from system storage")
			}
			if stat.Size() > 100*1024 {
				return nil, fmt.Errorf("template file %q is grather than 100KB", pth)
			}
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, errors.Wrapf(err, "read file %q", pth)
			}
			tmpl, err := template.New(pth).SetPath(pth).Parse(string(data))
			if err != nil {
				return nil, err
			}
			exc := tmpl.CreateExecutor().FuncsValues(ctx.FuncValues()).Funcs(template.FuncMap{
				"user_include": func(s *template.State, pth string, dot ...interface{}) {
					include(s.Writer(), pth, dot)
				},
			})
			exc.Context = ctx
			return exc, nil
		}

		if stat.Size() > 100*1024 {
			panic(fmt.Errorf("template file %q is grather than 100KB", pth))
		}

		yield := ctx.Yield
		defer func() {
			ctx.Yield = yield
		}()

		ctx.Yield = func(w io.Writer, results ...interface{}) {
			include(w, ctx.Request.URL.Path)
		}

		ctx.Execute("-", nil)
	} else if s, err := f.Stat(); err == nil {
		http.ServeContent(ctx.Writer, ctx.Request, ctx.Request.URL.Path[1:], s.ModTime(), f)
	} else {
		f.Close()
	}
}
