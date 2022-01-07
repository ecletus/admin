package admin

import (
	"net/http"

	"github.com/ecletus/responder"
	"github.com/moisespsena-go/httpu"
	"github.com/pkg/errors"

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
			cfg.FormAction = context.Resource.GetContextIndexURI(context, context.ParentResourceID...)
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
func (this *Controller) New(ctx *Context) {
	if ctx = ctx.ParentPreload(ParentPreloadNew); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	ctx.SetBasicType(NEW)
	ParseCreateConfig(ctx)

	ctx.Result = this.controller.(ControllerCreator).New(ctx)
	ctx.ResourceRecord = ctx.Result

	ctx.Execute("", this.controller.(ControllerCreator).New(ctx))
}

// Create create data
func (this *Controller) Create(ctx *Context) {
	if _, ok := this.controller.(ControllerCreator); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
	}
	if ctx = ctx.ParentPreload(ParentPreloadCreate); ctx.HasError() {
		ctx.LogErrors()
		return
	}
	ctx.SetBasicType(NEW)
	res := ctx.Resource

	if setuper, ok := this.controller.(ControllerSetuper); ok {
		ctx.AddError(setuper.SetupContext(ctx))
	}

	var record interface{}

	if !ctx.HasError() {
		record = this.controller.(ControllerCreator).New(ctx)

		if !ctx.HasError() {
			if ctx.AddError(res.Decode(ctx.Context, record)); !ctx.HasError() {
				this.controller.(ControllerCreator).Create(ctx, record)
			}
		}
	}

	ctx.Result = record
	ctx.ResourceRecord = record

	var (
		cfg             = ParseCreateConfig(ctx)
		messages        []string
		messageDisabled bool
	)

	if ctx.HasError() {
		if cfg.ErrorCallback != nil {
			cfg.ErrorCallback(ctx, ctx.Errors)
		}
		if ctx.Writer.WroteHeader() {
			return
		}
		responder.With("html", func() {
			ctx.Writer.WriteHeader(HTTPUnprocessableEntity)
			ctx.Execute("shared/errors", record)
		}).With([]string{"json", "xml"}, func() {
			ctx.Api = true
			ctx.Writer.WriteHeader(HTTPUnprocessableEntity)
			ctx.Encode(map[string]interface{}{"errors": ctx.GetErrors()})
		}).Respond(ctx.Request)
	} else {
		ctx.ResourceID = ctx.Resource.GetKey(ctx.Result)
		if ctx.Resource.Config.Wizard == nil {
			if ctx.Request.Header.Get("X-Flash-Messages-Disabled") != "true" {
				message := string(ctx.tt(I18NGROUP+".form.successfully_created",
					NewResourceRecorde(ctx, res, record),
					"{{.}} was successfully created"))

				if cfg.SuccessCallback != nil {
					cfg.SuccessCallback(ctx, message)
					if ctx.Writer.WroteHeader() {
						return
					}
				}

				messages = append(messages, message)
			}

			if cfg.RedirectTo == "" {
				cfg.RedirectTo = ctx.RedirectTo
			}
		}

		var respond = func(messages []string) {
			defer ctx.LogErrors()
			ctx.SetBasicType(SHOW)
			ctx.DefaulLayout()

			if ctx.Resource.Config.Wizard != nil {
				wz := record.(WizardModelInterface)
				ctx.Writer.Header().Set("X-Next-Step", wz.CurrentStepName())
				url := res.GetContextURI(ctx, res.GetKey(record))
				httpu.Redirect(ctx.Writer, ctx.Request, url, http.StatusSeeOther)
			} else {
				responder.With("html", func() {
					if !messageDisabled {
						ctx.FlashS("success", messages...)
					}
					url := cfg.RedirectTo
					if url == "" {
						if ctx.Request.URL.Query().Get("continue_editing") != "" || ctx.Request.URL.Query().Get("continue_editing_url") != "" {
							url = res.GetContextURI(ctx, res.GetKey(record)) + P_OBJ_UPDATE_FORM
						} else {
							url = res.GetContextIndexURI(ctx)
						}
					}
					httpu.Redirect(ctx.Writer, ctx.Request, url, http.StatusSeeOther)
				}).With([]string{"json", "xml"}, func() {
					ctx.Api = true
					if ctx.Request.URL.Query().Get("continue_editing") != "" || ctx.Request.URL.Query().Get("continue_editing_url") != "" {
						url := res.GetContextURI(ctx, res.GetKey(record)) + P_OBJ_UPDATE_FORM
						httpu.Redirect(ctx.Writer, ctx.Request, url, http.StatusSeeOther)
						return
					}
					ctx.Encode(record)
				}).XAccept().Respond(ctx.Request)
			}
		}

		if stateName := ctx.Request.PostForm.Get("QorCreateState"); stateName != "" {
			for _, state := range ctx.Resource.UpdateStates {
				if state.Name == stateName {
					err := state.Handler(ctx, &messages, func() {
						respond(messages)
					})
					if err != nil {
						panic(errors.Wrapf(err, "State %q handler", stateName))
					}
					return
				}
			}
		}
		respond(messages)
	}
}
