package admin

import (
	"net/http"

	"github.com/ecletus/core/resource"
	"github.com/ecletus/responder"

	"github.com/moisespsena-go/httpu"
)

func (this *Controller) LoadIndexData(ctx *Context) interface{} {
	if setuper, ok := this.controller.(ControllerSetuper); ok {
		if ctx.AddError(setuper.SetupContext(ctx)); ctx.HasError() {
			return nil
		}
	}
	return this.controller.(ControllerIndex).Index(ctx)
}

// Index render index page
func (this *Controller) Index(ctx *Context) {
	if _, ok := this.controller.(ControllerIndex); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
	}

	ctx.SetBasicType(INDEX)
	ctx.DefaulLayout()
	defer ctx.LogErrors()

	if ctx = ctx.ParentPreload(ParentPreloadIndex); ctx.HasError() {
		return
	}

	ctx.Context = ctx.SetValue(resource.AutoLoadLinkDisabled, true)

	responder.With("html", func() {
		var result interface{}
		if ctx.LoadDisplayOrError() {
			result = this.LoadIndexData(ctx)
		}

		if pagination := ctx.Searcher.Pagination; pagination.CurrentPage > 1 && pagination.CurrentPage > pagination.Pages {
			url, _ := ctx.PatchCurrentURL("page", pagination.Pages)
			httpu.Redirect(ctx.Writer, ctx.Request, url, http.StatusSeeOther)
			return
		}
		ctx.ResourceItems = result
		ctx.Result = result

		ctx.Execute("", result)
	}).With([]string{"json", "xml"}, func() {
		if ctx.ValidateLayoutOrError() {
			result := this.LoadIndexData(ctx)
			if ctx.HasError() {
				return
			}
			ctx.Api = true
			ctx.ResourceItems = result
			ctx.Encode(result)
		}
	}).Respond(ctx.Request)
}
