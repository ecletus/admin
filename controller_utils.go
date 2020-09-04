package admin

import (
	"net/http"

	"github.com/ecletus/responder"
)

// Show render show page
func (this *Controller) showOrEdit(context *Context, onRecorde func(record interface{}) bool) {
	context.DefaulLayout()

	if context.Type.Has(SHOW) {
		if !context.Resource.ReadOnly && !context.Resource.isSetShowAttrs {
			context.Type = context.Type.Clear(SHOW).Set(EDIT)
		}
	}

	responder.
		With("html", func() {
			if context.LoadDisplayOrError() {
				recorde := this.LoadShowData(context)
				if !context.HasError() {
					if recorde == nil {
						context.NotFound = true
						http.NotFound(context.Writer, context.Request)
						return
					}
					// if !context.Type.Has(DELETED) && context.Resource.IsSoftDeleted(recorde) {
					//	http.Redirect(context.Writer, context.Request, context.Resource.GetContextIndexURI(context.Context), http.StatusSeeOther)
					//	return
					// }
					if onRecorde(recorde) {
						context.Execute("", recorde)
					}
				} else {
					context.Execute("shared/errors", recorde)
				}
			}
		}).
		With([]string{"json", "xml"}, func() {
			if context.ValidateLayoutOrError() {
				recorde := this.LoadShowData(context)
				if !context.HasError() {
					if recorde == nil {
						context.NotFound = true
						http.NotFound(context.Writer, context.Request)
						return
					} else {
						if onRecorde(recorde) {
							context.Encode(recorde)
						}
					}
				} else {
					context.Writer.WriteHeader(http.StatusBadGateway)
					context.Writer.Write([]byte(context.Error()))
				}
			}
		}).
		XAccept().
		Respond(context.Request)
}
