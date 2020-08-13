package admin

func (h *RouteHandler) generateCrumbs(ctx *Context) {
	if h.CrumbsLoader != nil {
		rp := ctx.RouteContext.RoutePatterns
		// Skip admin path
		if ctx.Admin.Config.MountPath != "/" {
			rp = rp[1:]
		}
		h.CrumbsLoader.LoadCrumbs(h, ctx, rp...)
	}
}
