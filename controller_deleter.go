package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ecletus/responder"
	"github.com/ecletus/roles"
)

// Delete delete data
func (this *Controller) Delete(ctx *Context) {
	if _, ok := this.controller.(ControllerDeleter); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
		return
	}

	if ctx = ctx.ParentPreload(ParentPreloadDelete); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	res := ctx.Resource
	status := http.StatusOK

	if setuper, ok := this.controller.(ControllerSetuper); ok {
		ctx.AddError(setuper.SetupContext(ctx))
	}

	var (
		recorde interface{}
		msg     string
	)

	if !ctx.HasError() {
		ctx.AddError(ctx.CloneErr(func(ctx *Context) {
			recorde = this.controller.(ControllerReader).Read(ctx)
		}))
	}

	if recorde == nil && !ctx.HasError() {
		status = http.StatusNotFound
	} else {
		if ctx.HasError() {
			msg = string(ctx.tt(I18NGROUP+".form.failed_to_delete",
				NewResourceRecorde(ctx, res, recorde),
				"Failed to delete {{.}}"))
			status = http.StatusBadRequest
		} else {
			ctx.Result = recorde
			if !ctx.HasRecordPermission(ctx.Resource, recorde, roles.Delete) {
				msg = string(ctx.tt(I18NGROUP+".form.failed_to_delete",
					NewResourceRecorde(ctx, res, recorde),
					"Failed to delete {{.}}"))
				status = http.StatusBadRequest
			} else {
				this.controller.(ControllerDeleter).Delete(ctx, recorde)
				if ctx.HasError() {
					status = http.StatusBadRequest
					msg = string(ctx.tt(I18NGROUP+".form.failed_to_delete_error",
						map[string]interface{}{
							"Value": NewResourceRecorde(ctx, res, recorde),
							"Error": ctx.Errors,
						},
						"Failed to delete {{.Value}}: {{.Error}}"))
				} else {
					msg = string(ctx.tt(I18NGROUP+".form.successfully_deleted", NewResourceRecorde(ctx, res, recorde),
						"{{.}} was successfully deleted"))
				}
			}
		}
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			ctx.Flash(msg, "success")
			uri := ctx.Request.Header.Get("Referer")
			if uri == "" {
				uri = res.GetContextIndexURI(ctx)
			}
			if ctx.Request.Header.Get("X-Requested-With") != "" {
				ctx.Writer.Header().Set("X-Location", uri)
				ctx.Writer.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(ctx.Writer, ctx.Request, uri, http.StatusFound)
			}
			return
		} else {
			ctx.Writer.WriteHeader(status)
			ctx.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		ctx.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		} else {
			ctx.Flash(msg, "success")
		}
		ctx.Layout = "OK"
		ctx.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := ctx.Request.Header.Get("Referer")
		if uri == "" {
			uri = res.GetContextIndexURI(ctx)
		}
		http.Redirect(ctx.Writer, ctx.Request, uri, http.StatusFound)
	}).Respond(ctx.Request)
}

// BulkDelete delete many recordes
func (this *Controller) BulkDelete(context *Context) {
	context.Type = DELETE
	context.PermissionMode = roles.Delete

	if _, ok := this.controller.(ControllerBulkDeleter); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}

	var (
		res      = context.Resource
		status   = http.StatusOK
		recordes []interface{}
		keySep   = context.Request.URL.Query().Get("key_sep")
		keys     []string
		msg      string
	)

	if context.Request.ContentLength > 0 {
		if context.Request.Header.Get("Content-Type") == "application/json" {
			err := json.NewDecoder(context.Request.Body).Decode(&keys)
			if err != nil {
				status = http.StatusBadRequest
				context.AddError(err)
			}
		} else {
			keys = GetPrimaryValues(context.Request.Form)
		}
	} else {
		if keySep == "" {
			keySep = ":"
		}
		keys = strings.Split(context.Request.URL.Query().Get("key"), keySep)
	}

	if !context.HasError() {
		if setuper, ok := this.controller.(ControllerSetuper); ok {
			context.AddError(setuper.SetupContext(context))
		}

		if !context.HasError() {
			if res.Config.DeleteableDB != nil {
				context.SetRawDB(res.Config.DeleteableDB(context, context.DB()))
			}

			for _, key := range keys {
				if key == "" {
					continue
				}

				var err error
				if context.ResourceID, err = res.ParseID(key); err == nil {
					clone := context.Clone()
					record := this.controller.(ControllerReader).Read(clone)
					context.AddError(clone.Err())
					if !clone.HasError() && record != nil {
						recordes = append(recordes, record)
					}
				} else {
					context.AddError(err)
				}
			}
		}
	}

	context.ResourceID = nil

	if len(recordes) == 0 && !context.HasError() {
		status = http.StatusUnprocessableEntity
		msg = string(context.t(I18NGROUP+".records.deletion_empty",
			"No records to deletion"))
	} else {
		if context.HasError() {
			msg = string(context.tt(I18NGROUP+".form.failed_to_delete",
				NewResourceRecorde(context, res, recordes[len(recordes)-1]),
				"Failed to delete {{.}}"))
			status = http.StatusBadRequest
		} else {
			this.controller.(ControllerBulkDeleter).DeleteBulk(context, recordes...)
		}
	}

	if msg == "" && !context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.successfully_bulk_deleted", NewResourceRecorde(context, res, recordes...),
			"{{.}} was successfully deleted"))
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			uri := res.GetContextIndexURI(context, context.Context.ParentResourceID...)
			http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
		} else {
			context.Writer.WriteHeader(status)
			context.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		}
		if msg == "" && context.HasError() {
			msg = context.Error()
		}
		context.Layout = "OK"
		context.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := res.GetContextIndexURI(context, context.Context.ParentResourceID...)
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}
