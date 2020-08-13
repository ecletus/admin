package admin

import (
	"net/http"
)

// Controller admin controller
type Controller struct {
	action     *Action
	controller interface{}
}

// HTTPUnprocessableEntity error status code
const HTTPUnprocessableEntity = 422

// Search render search page
func (this *Controller) Search(context *Context) {
	if _, ok := this.controller.(ControllerSearcher); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}

	type Result struct {
		Context  *Context
		Resource *Resource
		Results  interface{}
	}

	var (
		searchResult = Result{Context: context, Resource: context.Resource}
		err          error
	)

	if searchResult.Results, err = context.FindMany(); err != nil {
		context.AddError(err)
	}

	context.Execute("search_center", []Result{searchResult})
}
