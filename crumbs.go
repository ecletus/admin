package admin

type CrumbsLoader interface {
	LoadCrumbs(rh *RouteHandler, ctx *Context, pattern ...string)
}
