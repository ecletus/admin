package admin

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/ecletus/responder"
	"github.com/jinzhu/copier"
	"github.com/moisespsena-go/httpu"

	"github.com/ecletus/core"
	"github.com/moisespsena-go/aorm"

	field_access "github.com/moisespsena-go/field-access"
	"github.com/moisespsena-go/valuesmap"

	"github.com/moisespsena/template/html/template"
)

func ParseUpdateConfig(context *Context) (cfg *UpdateConfig) {
	if tcfg := context.RouteHandler.Config.HandlerConfig; tcfg != nil {
		if opt, ok := tcfg.(*UpdateConfig); ok {
			cfg2 := *opt
			cfg = &cfg2
		}
	}

	if cfg == nil {
		cfg = &UpdateConfig{}
		cfg.Load(core.Configors{context.RouteHandler.Config, context.Resource}, context)
	}

	return
}

func beforeRender(context *Context, cfg *UpdateConfig) {
	if cfg.FormAction == "" {
		cfg.FormAction = context.Resource.GetContextURI(context.Context, context.ResourceID, context.ParentResourceID...)
	}

	if cfg.ButtonLabel == "" {
		cfg.ButtonLabel = context.Ts(I18NGROUP + ".form.save_changes")
	}

	context.Funcs(template.FuncMap{
		"update_config": func() *UpdateConfig {
			return cfg
		},
	})
}

// Edit render edit page
func (this *Controller) Edit(context *Context) {
	if _, ok := this.controller.(ControllerUpdater); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}
	context.Type = EDIT

	this.showOrEdit(context, func(record interface{}) bool {
		context.Result = record
		cfg := ParseUpdateConfig(context)
		if cfg.Prepare != nil {
			cfg.Prepare(context)
			if context.Writer.WroteHeader() {
				return false
			}
		}
		beforeRender(context, cfg)
		return true
	})

}

// Update update data
func (this *Controller) Update(context *Context) {
	if _, ok := this.controller.(ControllerUpdater); !ok {
		context.NotFound = true
		http.NotFound(context.Writer, context.Request)
	}

	context.Type = EDIT
	context.DefaulLayout()
	if !context.ValidateLayoutOrError() {
		return
	}
	var recorde, old interface{}
	res := context.Resource

	if !context.LoadDisplayOrError() {
		return
	}

	recorde = this.LoadShowData(context)

	var notLoad bool
	if check, ok := this.controller.(CheckLoaderForUpdater); ok {
		notLoad = !check.IsLoadForUpdate(context)
	}

	cfg := ParseUpdateConfig(context)

	if !context.HasError() {
		context.Result = recorde
		if cfg.Prepare != nil {
			cfg.Prepare(context)
			if context.Writer.WroteHeader() {
				return
			}
		}

		old = res.New()

		if err := copier.Copy(old, recorde); err != nil {
			context.AddError(err)
			goto done
		}

		decerr := res.NewDecoder(context.Context, recorde).
			SetNotLoad(true).
			Decode(recorde)

		if context.AddError(decerr); !context.HasError() {
			if !notLoad && old == nil {
				if reader, ok := this.controller.(ControllerReader); ok {
					context.WithDB(func(context *Context) {
						context.AddError(context.CloneErr(func(ctx *Context) {
							old = reader.Read(ctx)
						}))
					}, context.DB().Set(aorm.OptKeySkipPreload, true))
				}
			}
			this.controller.(ControllerUpdater).Update(context, recorde, old)
		}
	}

done:
	context.Result = recorde

	beforeRender(context, cfg)

	if context.HasError() {
		if cfg.ErrorCallback != nil {
			cfg.ErrorCallback(context, context.Errors)
			if context.Writer.WroteHeader() {
				return
			}
		}
		context.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			context.Execute("", recorde)
		}).With([]string{"json", "xml"}, func() {
			context.Encode(map[string]interface{}{"errors": context.GetErrors()})
		}).Respond(context.Request)
	} else {
		var messageDisabled bool
		if context.Resource.Config.Wizard != nil {
			wz := recorde.(WizardModelInterface)
			if wz.IsDone() {
				destRes := context.Resource.Config.Wizard.BaseResource
				dest := wz.GetDestination()
				context.Writer.Header().Set("X-Steps-Done", "true")
				message := string(context.tt(I18NGROUP+".form.successfully_created",
					NewResourceRecorde(context, destRes, dest), "{{.}} was successfully created"))
				if cfg.SuccessCallback != nil {
					cfg.SuccessCallback(context, old, recorde, &message)
					if context.Writer.WroteHeader() {
						return
					}
				}
				context.Flash(message, "success")

				url := cfg.RedirectTo
				if url == "" {
					if context.Request.URL.Query().Get("continue_editing") != "" || context.Request.URL.Query().Get("continue_editing_url") != "" {
						url = res.Config.Wizard.BaseResource.GetContextURI(context.Context, destRes.GetKey(dest)) + P_OBJ_UPDATE_FORM
					} else {
						url = res.Config.Wizard.BaseResource.GetContextIndexURI(context.Context)
					}
				}
				httpu.Redirect(context.Writer, context.Request, url, http.StatusSeeOther)
				return
			} else {
				context.Writer.Header().Set("X-Next-Step", wz.CurrentStepName())
				httpu.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusSeeOther)
				return
			}
		}
		var message string
		if !messageDisabled {
			message = string(context.tt(I18NGROUP+".form.successfully_updated", NewResourceRecorde(context, res, recorde),
				"{{.}} was successfully updated"))
			if cfg.SuccessCallback != nil {
				cfg.SuccessCallback(context, old, recorde, &message)
				if context.Writer.WroteHeader() {
					return
				}
			}
		}

		defer context.LogErrors()
		context.Type = SHOW
		context.DefaulLayout()
		responder.With("html", func() {
			if !messageDisabled {
				context.Flash(message, "success")
			}
			url := context.RedirectTo
			if url == "" {
				if url = cfg.RedirectTo; url == "" {
					if res.Config.Singleton {
						url = res.GetContextIndexURI(context.Context)
					} else {
						url = res.GetContextURI(context.Context, res.GetKey(recorde))
					}
					url += P_OBJ_UPDATE_FORM
				}
			}
			if context.Request.URL.Query().Get("continue_editing") != "" {
				http.Redirect(context.Writer, context.Request, url, http.StatusFound)
				return
			}
			httpu.Redirect(context.Writer, context.Request, url, http.StatusFound)
		}).With([]string{"json", "xml"}, func() {
			if context.Request.FormValue("qorInlineEdit") != "" {
				rresult := reflect.ValueOf(recorde)
				for rresult.Kind() == reflect.Ptr {
					rresult = rresult.Elem()
				}
				newResult := make(map[string]interface{})

				for key := range context.Request.Form {
					if strings.HasPrefix(key, "QorResource.") {
						key = strings.TrimPrefix(key, "QorResource.")
						f := rresult.FieldByName(key)
						if f.IsValid() {
							newResult[key] = f.Interface()
						} else if v, ok := field_access.Get(recorde, key); ok {
							newResult[key] = v
						}
					}
				}
				newResult = valuesmap.ParseMap(newResult)
				context.Encode(newResult)
				return
			}
			context.Encode(recorde)
		}).XAccept().Respond(context.Request)
	}
}
