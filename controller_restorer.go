package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ecletus/responder"

	"github.com/go-aorm/aorm"
)

// Delete delete data
func (this *Controller) DeletedIndex(context *Context) {
	if _, ok := this.controller.(ControllerRestorer); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}
	context.SetBasicType(INDEX | DELETED)
	context.DefaulLayout()

	if setuper, ok := this.controller.(ControllerSetuper); ok {
		context.AddError(setuper.SetupContext(context))
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.LoadDisplayOrError()
			context.Execute(context.Type.S(), nil)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
		return
	}

	ctrl := this.controller.(ControllerRestorer)
	defer context.LogErrors()
	responder.With("html", func() {
		var result interface{}
		if context.LoadDisplayOrError() {
			result = ctrl.DeletedIndex(context)
		}
		context.Execute(context.Type.S(), result)
	}).With([]string{"json", "xml"}, func() {
		if context.ValidateLayoutOrError() {
			result := ctrl.DeletedIndex(context)
			context.Api = true
			context.Encode(result)
		}
	}).Respond(context.Request)
}

// BulkDelete delete many recordes
func (this *Controller) Restore(context *Context) {
	if _, ok := this.controller.(ControllerRestorer); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}

	var (
		res    = context.Resource
		status = http.StatusOK
		keySep = context.Request.URL.Query().Get("key_sep")
		keys   []string
		msg    string
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

	if len(keys) == 0 {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		return
	}

	context.Type |= DELETED

	context.ResourceID = nil
prepare:
	if context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.failed_to_restore",
			NewResourceRecorde(context, res),
			"Failed to restore {{.}}"))
		status = http.StatusBadRequest
	} else {
		if setuper, ok := this.controller.(ControllerSetuper); ok {
			context.AddError(setuper.SetupContext(context))
		}
		if !context.HasError() {
			var ids = make([]aorm.ID, len(keys))
			for i, key := range keys {
				var err error
				if ids[i], err = res.ParseID(key); err != nil {
					context.AddError(err)
					goto prepare
				}
			}
			this.controller.(ControllerRestorer).Restore(context, ids...)
		}
	}

	if msg == "" && !context.HasError() {
		msg = string(context.tt(I18NGROUP+".form.successfully_restored", NewResourceRecorde(context, res),
			"{{.}} was successfully restored"))
	}

	responder.With("html", func() {
		if status == http.StatusOK {
			url := res.GetContextIndexURI(context)

			if context.Request.URL.Query().Get("continue_editing") != "" {
				http.Redirect(context.Writer, context.Request, url+"/"+keys[0], http.StatusFound)
				return
			} else if context.Request.URL.Query().Get("continue_editing_url") != "" {
				context.Writer.Header().Set("X-Continue-Editing-Url", url+"/"+keys[0])
				context.Writer.WriteHeader(http.StatusNoContent)
				return
			}

			http.Redirect(context.Writer, context.Request, url, http.StatusFound)
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
