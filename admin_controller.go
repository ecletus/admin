package admin

type AdminController struct {
	Admin            *Admin
	DashboardNamer   func(context *Context, defaul string) string
	DashboardHandler func(context *Context, name string)
	MenuIndexHandler func(context *Context, node *RouteNode)
}

// Dashboard render dashboard page
func (this *AdminController) MenuIndex(context *Context, node *RouteNode) {
	var (
		baseDir = "menu_index"
		name    = "default"
	)
	if context.Anonymous() {
		baseDir = AnonymousDirName + "/" + baseDir
	}
	name = baseDir + "/" + name
	if this.MenuIndexHandler != nil {
		this.MenuIndexHandler(context, node)
	} else {
		context.Execute(name, node)
	}
}

// Dashboard render dashboard page
func (this *AdminController) Dashboard(context *Context) {
	var (
		baseDir = "dashboard"
		name    = "default"
	)
	if context.Anonymous() {
		baseDir = AnonymousDirName + "/" + baseDir
	}
	name = baseDir + "/" + name
	if this.DashboardNamer != nil {
		name = this.DashboardNamer(context, name)
	}
	if this.DashboardHandler != nil {
		this.DashboardHandler(context, name)
	} else {
		context.Execute(name, nil)
	}
}

// SearchCenter render search center page
func (this *AdminController) SearchCenter(context *Context) {
	type Result struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}
	var (
		searchResults []Result
		err           error
	)

	for _, res := range context.GetSearchableResources() {
		var (
			resourceName = context.Request.URL.Query().Get("resource_name")
			ctx          = context.Clone().setResource(res)
			searchResult = Result{Context: ctx, Resource: res}
		)

		if resourceName == "" || res.ToParam() == resourceName {
			if searchResult.Results, err = ctx.ParseFindMany(); err != nil {
				context.AddError(err)
			}
		}
		searchResults = append(searchResults, searchResult)
	}
	context.Execute("search_center", searchResults)
}
