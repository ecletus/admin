package admin

import (
	"net/http"
	"strings"

	"github.com/ecletus/responder"
	"github.com/ecletus/roles"
)

// Show render show page
func (this *Controller) showOrEdit(context *Context, onRecorde func(record interface{}) bool) {
	context.DefaulLayout()

	if context.Type.Has(SHOW) {
		if !context.Resource.ReadOnly && !context.Resource.Sections.Default.Screen.Show.IsSetI() {
			context.Type = context.Type.Clear(SHOW).Set(EDIT)
			context.PermissionMode = roles.Update
		}
	}

	recorde := this.LoadShowData(context)

	if recorde == nil && !context.HasError() {
		context.Writer.WriteHeader(http.StatusNotFound)

		responder.
			With("html", func() {
				context.NotFound = true
				context.Execute("record_not_found", recorde)
			}).
			Respond(context.Request)
		return
	}

	if context.HasError() {
		context.Execute("shared/errors", nil)
		return
	}
	if context.Resource.AdminHasRecordPermission(context.PermissionMode, context, recorde).Deny() {
		if strings.HasSuffix(context.OriginalURL.Path, "/edit") {
			http.Redirect(context.Writer, context.Request, strings.TrimSuffix(context.OriginalURL.Path, "/edit"), http.StatusFound)
		} else {
			context.Writer.WriteHeader(http.StatusForbidden)
		}
		return
	}

	if context.Resource.IsSoftDeleted(recorde) {
		context.Type |= DELETED
	}

	responder.
		With("html", func() {
			if context.LoadDisplayOrError() {
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
