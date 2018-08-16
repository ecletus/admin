package admin

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/moisespsena/go-route"
	"github.com/moisespsena/go-topsort"
	"github.com/aghape/auth"
	"github.com/aghape/common"
	"github.com/aghape/aghape"
	"github.com/aghape/roles"
)

type Interseptor func(w http.ResponseWriter, req *http.Request, serv func(w http.ResponseWriter, req *http.Request))

// ServeHTTP dispatches the handler registered in the matched route
func (admin *Admin) routeInterseptor(chain *route.ChainHandler) {
	qorContext := qor.ContexFromRouteContext(chain.Context)
	staticURL := qorContext.StaticURL + "/admin"
	req, qorContext := qorContext.NewChild(nil, admin.Config.MountPath)
	qorContext.StaticURL = staticURL

	var (
		context      = admin.NewContext(qorContext)
		RelativePath = "/" + strings.Trim(strings.TrimPrefix(req.URL.Path, admin.Router.Prefix()), "/")
	)

	context.RouteContext = chain.Context

	SetContexToChain(chain, context)

	// Parse Request Form
	req.ParseMultipartForm(2 * 1024 * 1024)

	// Set Request Method
	if method := req.Form.Get("_method"); method != "" {
		req.Method = strings.ToUpper(method)
		chain.Context.RouteMethod = req.Method
	}

	if regexp.MustCompile("^/(assets|themes)/.*$").MatchString(RelativePath) && strings.ToUpper(req.Method) == "GET" {
		(&Controller{Admin: admin}).Asset(context)
		return
	}

	auth.InterceptFuncIfAuth(admin.Auth, chain.Writer, req, func(ok bool) {
		// Set Current User
		var currentUser common.User
		//var permissionMode roles.PermissionMode

		if ok {
			currentUser = admin.Auth.GetCurrentUser(context)
			context.CurrentUser = currentUser
			context.DB = context.DB.Set(PKG+".current_user", context.CurrentUser)
		}
		context.Roles = roles.MatchedRoles(req, currentUser)
		context.Breadcrumbs().Append(qor.NewBreadcrumb(context.GenURL(), I18NGROUP+".layout.title", ""))

		oldKey := chain.Context.DefaultValueKey
		defer func() {
			chain.Context.DefaultValueKey = oldKey
		}()
		chain.Context.DefaultValueKey = CONTEXT_KEY
		chain.Next(req)

		/*

			switch req.Method {
			case "GET":
				permissionMode = roles.Read
			case "PUT":
				permissionMode = roles.Update
			case "POST":
				permissionMode = roles.Create
			case "DELETE":
				permissionMode = roles.Delete
			}

			handlers := admin.Router.routers[strings.ToUpper(req.Method)]
			for _, handler := range handlers {
				if params, _, ok := utils.ParamsMatch(handler.Path, RelativePath); ok && handler.HasPermission(permissionMode, context.Context) {
					if params.Size > 0 {
						req.URL.RawQuery = url.Values(params.Dict()).Encode() + "&" + req.URL.RawQuery
					}

					context.RouteHandler = handler
					context.setResource(handler.Config.Resource)

					if context.Resource == nil {
						if matches := regexp.MustCompile(path.Join(admin.router.Prefix, `([^/]+)`)).FindStringSubmatch(req.URL.Path); len(matches) > 1 {
							context.setResource(admin.GetResourceByID(matches[1]))
						}
					}

					if context.Resource != nil {
						context.ParentResourceID = context.Resource.ParentsID(params)
						pres := context.Resource.ParentResource

						for i := len(context.ParentResourceID); i > 0; i-- {
							basicValue, errpr := pres.FindOneBasicHandler(context.DB, context.ParentResourceID[i])
							puri := pres.GetIndexURI(context, context.ParentResourceID...)
							context.Breadcrumbs().Append(&qor.NewBreadcrumb(puri, basicValue.Label(), ""))
						}
					}
					break
				}
			}
		*/
	})
}

func (admin *Admin) handlerInterseptor(chain *route.ChainHandler) {
	context := ContextFromChain(chain)
	if context.Resource != nil {
		context = context.Admin.NewContextForResource(context, context.Resource)
		SetContexToChain(chain, context)
	}
	context.RouteContext = chain.Context

	if h, ok := chain.Endpoint.(*RouteHandler); ok {
		context.PermissionMode = h.Config.PermissionMode
	} else {
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
		//if h.HasPermission(permissionMode, context.Context) {
		//}
	}

	chain.Pass()
}

// NewServeMux generate http.Handler for admin
func (admin *Admin) NewServeMux(name ...string) *route.Mux {
	// Register default routes & middlewares
	router := admin.Router
	if len(name) > 0 {
		router.Name = name[0]
	}
	adminController := &Controller{Admin: admin}
	router.Get("/", adminController.Dashboard)
	router.Get("/search", adminController.SearchCenter)

	browserUserAgentRegexp := regexp.MustCompile("Mozilla|Gecko|WebKit|MSIE|Opera")
	router.Use(&route.Middleware{
		Name: PKG + ".csrf_check",
		Handler: func(chain *route.ChainHandler) {
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

	router.Use(&route.Middleware{
		Name: PKG + ".handler",
		Handler: func(chain *route.ChainHandler) {
			chain.Writer.Header().Set("Cache-control", "no-store")
			chain.Writer.Header().Set("Pragma", "no-cache")
			chain.Pass()
		},
	})

	router.HandlerIntersept(&route.Middleware{
		Name:    PKG + ".handler_interseptor",
		Handler: admin.handlerInterseptor,
	})

	router.HandlerIntersept(
		&route.Middleware{
			Name: PKG + ".resource.main",
			Handler: func(chain *route.ChainHandler) {
				// skip the /admin/* pattern
				patterns := chain.Context.RoutePatterns[1:]

				if len(patterns) < 2 {
					chain.Pass()
					return
				}

				resourceParam := strings.TrimSuffix(patterns[0][1:], "/*")
				res := admin.GetResourceByParam(resourceParam)
				if res == nil {
					chain.Pass()
					return
				}

				for i, l := 1, len(patterns); i < l; i += 2 {
					// id pattern
					idPattern := "/" + res.ParamIDPattern() + "/*"
					pattern := patterns[i]
					if pattern != idPattern {
						break
					}
					resourceParam := strings.TrimSuffix(patterns[i+1][1:], "/*")
					subRes := res.GetResourceByParam(resourceParam)
					if subRes != nil {
						res = subRes
					}
				}

				context := ContextFromChain(chain)
				context.setResource(res)

				if context.URLParams().Size > 0 {
					var parentIds []string
					var parents []*Resource
					p := res.ParentResource

					for p != nil {
						parents = append(parents, p)
						key := p.ParamIDName()
						id := context.URLParam(key)
						parentIds = append(parentIds, id)
						p = p.ParentResource
					}

					context.ResourceID = context.URLParam(res.ParamIDName())
					db := context.Top().DB
					var ids []string
					for i, id := range parentIds {
						p = parents[i]
						uri := p.GetContextIndexURI(context.Context, ids...)
						context.Breadcrumbs().Append(qor.NewBreadcrumb(uri, p.GetLabelKey(true), ""))
						model, err := p.FindOneBasic(db, id)

						if err != nil {
							panic(err)
						}

						uri = p.GetContextURI(context.Context, id, ids...)
						context.Breadcrumbs().Append(qor.NewBreadcrumb(uri, model.BasicLabel(), model.BasicIcon()))
						ids = append(ids, id)
					}

					if context.ResourceID != "" {
						uri := res.GetContextIndexURI(context.Context, ids...)
						context.Breadcrumbs().Append(qor.NewBreadcrumb(uri, res.GetLabelKey(true), ""))
					}

					topsort.Reverse(parentIds)
					context.ParentResourceID = parentIds
				}

				chain.Pass()
			},
		})

	return admin.InitRoutes()
}
