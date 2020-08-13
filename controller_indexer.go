package admin

import (
	"net/http"

	"github.com/ecletus/responder"
)

func (this *Controller) LoadIndexData(context *Context) interface{} {
	if setuper, ok := this.controller.(ControllerSetuper); ok {
		if context.AddError(setuper.SetupContext(context)); context.HasError() {
			return nil
		}
	}
	return this.controller.(ControllerIndex).Index(context)
}

// Index render index page
func (this *Controller) Index(context *Context) {
	if _, ok := this.controller.(ControllerIndex); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}
	context.Type = INDEX
	context.DefaulLayout()
	defer context.LogErrors()
	responder.With("html", func() {
		var result interface{}
		if context.LoadDisplayOrError() {
			result = this.LoadIndexData(context)
		}
		context.Execute("", result)
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result := this.LoadIndexData(context)
			context.Api = true
			context.Encode(result)
		}
	}).Respond(context.Request)
}
