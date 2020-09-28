package admin

import (
	"net/http"

	"github.com/ecletus/responder"
	"github.com/moisespsena-go/httpu"

	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core"
)

func ParseCreateConfig(context *Context) (cfg *CreateConfig) {
	if tcfg := context.RouteHandler.Config.HandlerConfig; tcfg != nil {
		if opt, ok := tcfg.(*CreateConfig); ok {
			cfg2 := *opt
			cfg = &cfg2
		}
	}

	if cfg == nil {
		cfg = &CreateConfig{}
		cfg.Load(core.Configors{context.RouteHandler.Config, context.Resource}, context)

		if cfg.ButtonLabel == "" {
			cfg.ButtonLabel = context.Ts(I18NGROUP + ".form.add")
		}

		if cfg.FormAction == "" {
			cfg.FormAction = context.Resource.GetContextIndexURI(context.Context, context.ParentResourceID...)
		}
	}

	context.Funcs(template.FuncMap{
		"create_config": func() *CreateConfig {
			return cfg
		},
	})

	return
}

// New render new page
func (this *Controller) New(context *Context) {
	context.Type = NEW
	ParseCreateConfig(context)
	context.Execute("", this.controller.(ControllerCreator).New(context))
}

// Create create data
func (this *Controller) Create(context *Context) {
	if _, ok := this.controller.(ControllerCreator); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}
	context.Type = NEW
	res := context.Resource

	if setuper, ok := this.controller.(ControllerSetuper); ok {
		context.AddError(setuper.SetupContext(context))
	}

	var recorde interface{}

	if !context.HasError() {
		recorde = this.controller.(ControllerCreator).New(context)

		if !context.HasError() {
			if context.AddError(res.Decode(context.Context, recorde)); !context.HasError() {
				this.controller.(ControllerCreator).Create(context, recorde)
			}
		}
	}

	context.Result = recorde
	cfg := ParseCreateConfig(context)
	if context.HasError() {
		if cfg.ErrorCallback != nil {
			cfg.ErrorCallback(context, context.Errors)
		}
		if context.Writer.WroteHeader() {
			return
		}
		responder.With("html", func() {
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Execute("shared/errors", recorde)
		}).With([]string{"json", "xml"}, func() {
			context.Api = true
			context.Writer.WriteHeader(HTTPUnprocessableEntity)
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		context.ResourceID = context.Resource.GetKey(context.Result)
		if context.Resource.Config.Wizard == nil {
			if context.Request.Header.Get("X-Flash-Messages-Disabled") != "true" {
				message := string(context.tt(I18NGROUP+".form.successfully_created",
					NewResourceRecorde(context, res, recorde),
					"{{.}} was successfully created"))

				if cfg.SuccessCallback != nil {
					cfg.SuccessCallback(context, message)
					if context.Writer.WroteHeader() {
						return
					}
				}

				context.Flash(message, "success")
			}

			if cfg.RedirectTo == "" {
				cfg.RedirectTo = context.RedirectTo
			}
		}

		defer context.LogErrors()
		context.Type = SHOW
		context.DefaulLayout()

		if context.Resource.Config.Wizard != nil {
			wz := recorde.(WizardModelInterface)
			context.Writer.Header().Set("X-Next-Step", wz.CurrentStepName())
			url := res.GetContextURI(context.Context, res.GetKey(recorde))
			httpu.Redirect(context.Writer, context.Request, url, http.StatusSeeOther)
		} else {
			responder.With("html", func() {
				url := cfg.RedirectTo
				if url == "" {
					if context.Request.URL.Query().Get("continue_editing") != "" || context.Request.URL.Query().Get("continue_editing_url") != "" {
						url = res.GetContextURI(context.Context, res.GetKey(recorde)) + P_OBJ_UPDATE_FORM
					} else {
						url = res.GetContextIndexURI(context.Context)
					}
				}
				httpu.Redirect(context.Writer, context.Request, url, http.StatusSeeOther)
			}).With([]string{"json", "xml"}, func() {
				context.Api = true
				if context.Request.URL.Query().Get("continue_editing") != "" || context.Request.URL.Query().Get("continue_editing_url") != "" {
					url := res.GetContextURI(context.Context, res.GetKey(recorde)) + P_OBJ_UPDATE_FORM
					httpu.Redirect(context.Writer, context.Request, url, http.StatusSeeOther)
					return
				}
				context.Encode(recorde)
			}).XAccept().Respond(context.Request)
		}
	}
}
