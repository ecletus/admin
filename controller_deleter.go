package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ecletus/responder"
)

// Delete delete data
func (this *Controller) Delete(context *Context) {
	if _, ok := this.controller.(ControllerDeleter); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
		return
	}
	res := context.Resource
	status := http.StatusOK

	if setuper, ok := this.controller.(ControllerSetuper); ok {
		context.AddError(setuper.SetupContext(context))
	}

	var (
		recorde interface{}
		msg     string
	)

	if !context.HasError() {
		context.AddError(context.CloneErr(func(ctx *Context) {
			recorde = this.controller.(ControllerReader).Read(ctx)
		}))
	}

	if recorde == nil && !context.HasError() {
		status = http.StatusNotFound
	} else {
		if context.HasError() {
			msg = string(context.tt(I18NGROUP+".form.failed_to_delete",
				NewResourceRecorde(context, res, recorde),
				"Failed to delete {{.}}"))
			status = http.StatusBadRequest
		} else {
			this.controller.(ControllerDeleter).Delete(context, recorde)
			if context.HasError() {
				status = http.StatusBadRequest
				msg = string(context.tt(I18NGROUP+".form.failed_to_delete_error",
					map[string]interface{}{
						"Value": NewResourceRecorde(context, res, recorde),
						"error": context.Errors,
					},
					"Failed to delete {{.Value}}: {{.error}}"))
			} else {
				msg = string(context.tt(I18NGROUP+".form.successfully_deleted", NewResourceRecorde(context, res, recorde),
					"{{.}} was successfully deleted"))
			}
		}
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			context.Flash(msg, "success")
			uri := context.Request.Header.Get("Referer")
			if uri == "" {
				uri = res.GetContextIndexURI(context.Context)
			}
			if context.Request.Header.Get("X-Requested-With") != "" {
				context.Writer.Header().Set("X-Location", uri)
				context.Writer.WriteHeader(http.StatusOK)
			} else {
				http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
			}
			return
		} else {
			context.Writer.WriteHeader(status)
			context.Writer.Write([]byte(msg))
		}
	}).With([]string{"json", "xml"}, func() {
		context.Writer.WriteHeader(status)
		messageStatus := "ok"
		if status != http.StatusOK {
			messageStatus = "error"
		} else {
			context.Flash(msg, "success")
		}
		context.Layout = "OK"
		context.Encode(map[string]string{"message": msg, "status": messageStatus})
		uri := context.Request.Header.Get("Referer")
		if uri == "" {
			uri = res.GetContextIndexURI(context.Context)
		}
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}

// BulkDelete delete many recordes
func (this *Controller) BulkDelete(context *Context) {
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
			keys = context.Request.Form["primary_values[]"]
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
			for _, key := range keys {
				if key == "" {
					continue
				}

				var err error
				if context.ResourceID, err = res.ParseID(key); err == nil {
					clone := context.Clone()
					recorde := this.controller.(ControllerReader).Read(clone)
					context.AddError(clone.Err())
					if !clone.HasError() {
						recordes = append(recordes, recorde)
					}
				} else {
					context.AddError(err)
				}
			}
		}
	}

	context.ResourceID = nil

	if len(recordes) == 0 && !context.HasError() {
		status = http.StatusNotFound
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
			uri := res.GetContextIndexURI(context.Context, context.Context.ParentResourceID...)
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
		uri := res.GetContextIndexURI(context.Context, context.Context.ParentResourceID...)
		http.Redirect(context.Writer, context.Request, uri, http.StatusFound)
	}).Respond(context.Request)
}
