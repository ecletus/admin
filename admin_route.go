package admin

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aghape/auth"
	"github.com/aghape/core"
	"github.com/aghape/roles"
	"github.com/moisespsena-go/xroute"
)

type Interseptor func(w http.ResponseWriter, req *http.Request, serv func(w http.ResponseWriter, req *http.Request))

// ServeHTTP dispatches the handler registered in the matched route
func (admin *Admin) routeInterseptor(chain *xroute.ChainHandler) {
	qorContext := core.ContexFromRouteContext(chain.Context)
	staticURL := qorContext.StaticURL + "/admin"
	req, qorContext := qorContext.NewChild(nil, admin.Config.MountPath)
	qorContext.StaticURL = staticURL

	var (
		context      = admin.NewContext(qorContext)
		RelativePath = "/" + strings.Trim(strings.TrimPrefix(req.URL.Path, admin.Config.MountPath), "/")
	)

	context.RouteContext = chain.Context

	SetContexToChain(chain, context)

	switch req.Method {
	case http.MethodDelete, http.MethodPost, http.MethodPut:
		// Parse Request Form
		req.ParseMultipartForm(2 * 1024 * 1024)

		// Set Request Method
		if method := req.Form.Get("_method"); method != "" {
			req.Method = strings.ToUpper(method)
			chain.Context.RouteMethod = req.Method
		}
	}

	if regexp.MustCompile("^/(assets|themes)/.*$").MatchString(RelativePath) && strings.ToUpper(req.Method) == "GET" {
		(&Controller{Admin: admin}).Asset(context)
		return
	}

	auth.InterceptFuncIfAuth(admin.Auth, chain.Writer, req, func(ok bool) {
		// Set Current User
		var currentUser = context.CurrentUser()
		//var permissionMode roles.PermissionMode

		if ok {
			context.DB = context.DB.SetCurrentUser(context.CurrentUser())
		}
		context.Roles = roles.MatchedRoles(req, currentUser)
		context.Breadcrumbs().Append(core.NewBreadcrumb(context.GenURL(), I18NGROUP+".layout.title", ""))

		oldKey := chain.Context.DefaultValueKey
		defer func() {
			chain.Context.DefaultValueKey = oldKey
		}()
		chain.Context.DefaultValueKey = CONTEXT_KEY
		chain.Next(req)
	})
}

func (admin *Admin) handlerInterseptor(chain *xroute.ChainHandler) {
	context := ContextFromChain(chain)
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

func RouteContextHandler(handler func(ctx *Context)) func(rctx *xroute.RouteContext) {
	return func(rctx *xroute.RouteContext) {
		handler(ContextFromRouteContext(rctx))
	}
}

func (admin *Admin) InitRoutes(router xroute.Router) {
	router.Intersept(&xroute.Middleware{
		Name:    PKG,
		Handler: admin.routeInterseptor,
	})

	adminController := &Controller{Admin: admin}
	router.Get("/", RouteContextHandler(adminController.Dashboard))
	router.Get("/search", RouteContextHandler(adminController.SearchCenter))

	browserUserAgentRegexp := regexp.MustCompile("Mozilla|Gecko|WebKit|MSIE|Opera")
	router.Use(&xroute.Middleware{
		Name: PKG + ".csrf_check",
		Handler: func(chain *xroute.ChainHandler) {
			request := chain.Request()
			if request.Method != "GET" {
				if browserUserAgentRegexp.MatchString(request.UserAgent()) {
					if referrer := request.Referer(); referrer != "" {
						if r, err := url.Parse(referrer); err == nil {
							if r.Host == request.Host {
								chain.Pass()
								return
							}
						}
					}
					chain.Writer.Write([]byte("Could not authorize you because 'CSRF detected'"))
					return
				}
			}

			chain.Pass()
		},
	})

	router.Use(&xroute.Middleware{
		Name: PKG + ".handler",
		Handler: func(chain *xroute.ChainHandler) {
			chain.Writer.Header().Set("Cache-control", "no-store")
			chain.Writer.Header().Set("Pragma", "no-cache")
			chain.Pass()
		},
	})

	router.HandlerIntersept(&xroute.Middleware{
		Name:    PKG + ".handler_interseptor",
		Handler: admin.handlerInterseptor,
	})

	for param, res := range admin.ResourcesByParam {
		pattern := "/" + param
		r := res.InitRoutes()
		router.Mount(pattern, r)
	}

	for _, r := range admin.onRouter {
		r(router)
	}
}
