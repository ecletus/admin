package admin

func (h *RouteHandler) generateCrumbs(ctx *Context) {
	if h.CrumbsLoader != nil {
		// Skip admin path
		rp := ctx.RouteContext.RoutePatterns[1:]
		h.CrumbsLoader.LoadCrumbs(h, ctx, rp...)
	}
}
