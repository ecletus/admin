package admin

import (
	"net/http"
	"strings"

	"github.com/ecletus/responder"

	"github.com/moisespsena-go/aorm"
)

type ActionController interface {
	Action(context *Context)
}

type ActionControl struct {
	action *Action
	Setup  func(ctx *Context) error
}

// Action handle action related requests
func (this *ActionControl) Action(context *Context) {
	var action = this.action

	if action.Available != nil {
		if !action.Available(context) {
			context.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	defer context.LogErrors()

	if this.Setup != nil {
		context.AddError(this.Setup(context))
	}

	if context.HasError() {
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("action", action)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(map[string]interface{}{"error": strings.Join(context.GetErrorsTS(), "; "), "status": "error"})
		}).Respond(context.Request)
		return
	}

	context.Type = EDIT | ACTION

	var actionArgument = &ActionArgument{
		Context: context,
		Action: action,
	}

	if action.Resource != nil {
		actionArgument.Argument = action.Resource.NewStruct(context.Context.Site)
		var err error

		if !action.BaseResource.IsSingleton() {
			values := context.Request.Form["primary_values[]"]
			actionArgument.PrimaryValues = make([]aorm.ID, len(values))
			for i, value := range values {
				if actionArgument.PrimaryValues[i], err = action.BaseResource.ParseID(value); err != nil {
					break
				}
			}
		}

		if err == nil && action.SetupArgument != nil {
			err = action.SetupArgument(actionArgument)
		}
		context.AddError(err)
	}

	if context.Request.Method == "GET" {
		if action.ShowHandler != nil {
			action.ShowHandler(actionArgument)
		}
		if action.ReadOnly {
			context.Type = SHOW | ACTION
			context.Execute("action_show", actionArgument)
		} else {
			context.Execute("action", actionArgument)
		}
	} else {
		var err error
		if context.HasError() {
			goto done
		}
		if action.Resource != nil {
			if err = action.Resource.Decode(context.Context, actionArgument.Argument, true); err != nil {
				goto done
			}
		}

		err = action.Handler(actionArgument)

	done:
		if !actionArgument.Context.Writer.WroteHeader() {
			if err == nil {
				message := actionArgument.successMessage
				if message == "" {
					message = string(context.tt(I18NGROUP+".actions.executed_successfully", action, "Action {{.Name}}: Executed successfully"))
				}
				if context.Request.Header.Get("X-Disabled-Success-Redirection") == "true" {
					context.Writer.Header().Set("X-Location", context.Request.Referer())
					context.Flash(message, "success")
					context.Writer.WriteHeader(http.StatusNoContent)
					return
				}
				responder.With("html", func() {
					context.Flash(message, "success")
					http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
				}).With([]string{"json"}, func() {
					context.Layout = "OK"
					context.Encode(map[string]string{"message": message, "status": "ok"})
				}).Respond(context.Request)
			} else {
				context.Writer.WriteHeader(HTTPUnprocessableEntity)
				responder.With("html", func() {
					context.AddError(err)
					context.Execute("action", actionArgument)
				}).With([]string{"json", "xml"}, func() {
					context.Layout = "OK"
					message := string(context.tt(I18NGROUP+".actions.executed_failed", action, "Action {{.Name}}: Failed to execute"))
					context.Encode(map[string]string{"error": message, "status": "error"})
				}).Respond(context.Request)
			}
		}
	}
}
