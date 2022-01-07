package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/responder"
	"github.com/moisespsena-go/aorm"
	"github.com/moisespsena-go/httpu"

	"github.com/ecletus/core/resource"
)

type ActionController interface {
	Action(context *Context)
}

type ActionControl struct {
	action *Action
	Setup  func(ctx *Context) error
}

// Action handle action related requests
func (this *ActionControl) Action(ctx *Context) {
	var action = this.action
	ctx.ResourceAction = action

	if ctx = ctx.ParentPreload(ParentPreloadAction); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	ctx.SetBasicType(ACTION)

	if action.Available != nil {
		if !action.Available(ctx) {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var (
		arg = &ActionArgument{
			Context: ctx,
			Action:  action,
		}

		parsePks = func() (err error) {
			if !action.BaseResource.IsSingleton() {
				var values []string
				switch ctx.Request.Method {
				case http.MethodGet:
					values = GetPrimaryValues(ctx.Request.URL.Query())
				case http.MethodPost, http.MethodPut:
					values = GetPrimaryValues(ctx.Request.Form)
				}

				if len(values) > 0 {
					arg.PrimaryValues = make([]aorm.ID, len(values))
					for i, value := range values {
						if arg.PrimaryValues[i], err = action.BaseResource.ParseID(value); err != nil {
							break
						}
					}
				}
			}
			return
		}
	)

	defer ctx.LogErrors()

	if ctx.ResourceID != nil {
		if !ctx.ResourceID.IsZero() && !action.SkipRecordLoad || action.RecordAvailable != nil {
			arg.Record = action.BaseResource.New()
			ctx.ResourceID.SetTo(arg.Record)
			var (
				err      error
				searcher = ctx.Clone().NewSearcher()
			)

			if action.FindRecord != nil {
				arg.Record, err = action.FindRecord(searcher)
			} else {
				arg.Record, err = searcher.FindOne()
			}
			ctx.AddError(err)
		}
		if !ctx.HasError() && action.RecordAvailable != nil {
			if !action.RecordAvailable(arg.Record, ctx) {
				ctx.Writer.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}

	if this.Setup != nil {
		ctx.AddError(this.Setup(ctx))
	}

	title := string(ctx.Tt(action))
	ctx.Breadcrumbs().Append(core.NewBreadcrumb("", title))
	ctx.PageTitle = title

	if formatter := GetOptActionTitleFormatter(ctx.Resource); formatter != nil {
		ctx.PageTitle = formatter(arg, ctx.PageTitle)
	} else if arg.Record != nil {
		ctx.PageTitle += " | " + ctx.Resource.GetLabel(ctx, false)
		if !ctx.Resource.Config.Singleton {
			ctx.PageTitle += " #" + ctx.Resource.GetKey(arg.Record).String()
		}
	}

	if ctx.HasError() {
		if ctx.Request.Method == action.Method && action.Method != http.MethodGet {
			ctx.Writer.WriteHeader(HTTPUnprocessableEntity)
		}
		responder.With("html", func() {
			ctx.Execute("action", arg)
			ctx.Errors.Reset()
		}).With([]string{"json", "xml"}, func() {
			ctx.Encode(map[string]interface{}{"error": strings.Join(ctx.GetErrorsTS(), "; "), "status": "error"})
		}).Respond(ctx.Request)
		return
	}

	var templateName string
	switch action.FormType {
	case ActionFormShow:
		ctx.SetBasicType(SHOW | ACTION)
		templateName = "action_show"
	case 0, ActionFormEdit:
		ctx.SetBasicType(EDIT | ACTION)
		templateName = "action_edit"
	case ActionFormNew:
		ctx.SetBasicType(NEW | ACTION)
		templateName = "action_new"
	}
	ctx.Result = arg

	if action.Resource != nil {
		arg.Argument = action.Resource.NewStruct(ctx.Context.Site)
		ctx.ResourceRecord = arg.Argument
		defer ctx.PushI18nGroup(action.Resource.i18nKey)()

		var err = parsePks()

		if err == nil && action.SetupArgument != nil {
			err = action.SetupArgument(arg)
		}
		ctx.AddError(err)
	} else if err := parsePks(); err != nil {
		ctx.AddError(err)
	} else {
		defer ctx.PushI18nGroup(ctx.Resource.i18nKey)()
	}

	var autoReload bool

	if !action.EmptyBulkAllowed && (action.ReadOnly() || ctx.Request.Method == "GET") {
		if !ctx.HasError() && action.ShowHandler != nil {
			ctx.AddError(action.ShowHandler(arg))
		}
		if ctx.HasError() {
			ctx.Execute("messages", ctx)
			return
		}

		ctx.Execute(templateName, arg)
		return
	} else {
		autoReload = true
	}

	var (
		err error
	)

	if ctx.HasError() {
		goto done
	}
	if action.Resource != nil {
		if err = action.Resource.Decode(ctx.Context, arg.Argument, resource.ProcSkipLoad); err != nil {
			goto done
		}
	}

	for _, state := range action.States {
		if val := ctx.Request.FormValue("QorActionState." + state.Name); val != "" {
			err = state.Handler(arg, val, func() {
				ctx.Type |= RE_RENDER
			})
			if ctx.Type.Has(RE_RENDER) && err == nil && !ctx.HasError() {
				ctx.Writer.Header().Set("X-Frame-Reload", "render-body")
				ctx.RequestLayout = "lite"
				ctx.Execute(templateName, arg)
				return
			}
			goto done
		}
	}

	err = action.Handler(arg)

done:

	if !arg.Context.Writer.WroteHeader() {
		if err == nil && !ctx.HasError() {
			message := arg.successMessage
			if message == "" {
				message = fmt.Sprintf(ctx.Ts(I18NGROUP+".actions.executed_successfully", "Action \"%s\": Executed successfully"), ctx.Tt(action))
			}
			if ctx.Request.Header.Get("X-Disabled-Success-Redirection") == "true" {
				ctx.Writer.Header().Set("X-Location", ctx.Request.Referer())
				ctx.Flash(message, "success")
				ctx.Writer.WriteHeader(http.StatusNoContent)
				return
			}
			responder.With("html", func() {
				ctx.Flash(message, "success")
				toUrl := ctx.Request.Referer()
				if action.RefreshURL != nil {
					toUrl = action.RefreshURL(arg.Record, ctx)
				}
				httpu.Redirect(ctx.Writer, ctx.Request, toUrl, http.StatusOK, autoReload)
			}).With([]string{"json"}, func() {
				ctx.Layout = "OK"
				ctx.Encode(map[string]string{"message": message, "status": "ok"})
			}).Respond(ctx.Request)
		} else {
			ctx.Writer.WriteHeader(HTTPUnprocessableEntity)
			responder.With("html", func() {
				ctx.AddError(err)
				if ctx.Request.Header.Get("X-Error-Body") == "true" {
					ctx.Writer.Write([]byte(strings.Join(ctx.GetErrorsTS(), "<br />")))
					return
				}
				ctx.Execute(templateName, arg)
			}).With([]string{"json", "xml"}, func() {
				ctx.Layout = "OK"
				if ctx.HasError() {
					ctx.AddError(err)
					ctx.Encode(map[string]interface{}{"errors": ctx.GetCleanFormattedErrors(), "status": "error"})
				} else {
					message := fmt.Sprintf(ctx.Ts(I18NGROUP+".actions.executed_failed", "Action %q: Failed to execute: %s"), ctx.Tt(action), ctx.ErrorTS(err))
					ctx.Encode(map[string]string{"error": message, "status": "error"})
				}
			}).Respond(ctx.Request)
		}
	}
}
