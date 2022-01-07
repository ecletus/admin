package admin

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/ecletus/core"
	"github.com/ecletus/responder"
	"github.com/ecletus/roles"
	"github.com/ecletus/validations"
	"github.com/jinzhu/copier"
	"github.com/moisespsena-go/aorm"
	field_access "github.com/moisespsena-go/field-access"
	"github.com/moisespsena-go/httpu"
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
		cfg.FormAction = context.Resource.GetContextURI(context, context.ResourceID, context.ParentResourceID...)
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
func (this *Controller) Edit(ctx *Context) {
	if _, ok := this.controller.(ControllerUpdater); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
	}

	if ctx = ctx.ParentPreload(ParentPreloadEdit); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	ctx.SetBasicType(EDIT)
	ctx.PermissionMode = roles.Update

	this.showOrEdit(ctx, func(record interface{}) bool {
		ctx.Result = record
		ctx.ResourceRecord = record

		cfg := ParseUpdateConfig(ctx)

		if cfg.Prepare != nil {
			cfg.Prepare(ctx)
			if ctx.Writer.WroteHeader() {
				return false
			}
		}
		beforeRender(ctx, cfg)
		return true
	})

}

// Update update data
func (this *Controller) Update(ctx *Context) {
	if _, ok := this.controller.(ControllerUpdater); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
	}

	if ctx = ctx.ParentPreload(ParentPreloadUpdate); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	ctx.SetBasicType(EDIT)
	ctx.DefaulLayout()
	if !ctx.ValidateLayoutOrError() {
		return
	}
	var record, old interface{}
	res := ctx.Resource

	if !ctx.LoadDisplayOrError() {
		return
	}

	record = this.LoadShowData(ctx)

	var notLoad bool
	if check, ok := this.controller.(CheckLoaderForUpdater); ok {
		notLoad = !check.IsLoadForUpdate(ctx)
	}

	cfg := ParseUpdateConfig(ctx)

	var (
		messageDisabled bool
		messages        []string
		respond         = func(messages []string) {
			ctx.SetBasicType(SHOW)
			ctx.DefaulLayout()
			responder.With("html", func() {
				if !messageDisabled {
					ctx.FlashS("success", messages...)
				}
				url := ctx.RedirectTo
				if url == "" {
					if url = cfg.RedirectTo; url == "" {
						if res.Config.Singleton {
							url = res.GetContextIndexURI(ctx)
						} else {
							url = res.GetContextURI(ctx, res.GetKey(record))
						}
						url += P_OBJ_UPDATE_FORM
					}
				}
				if q := ctx.Request.URL.Query(); q.Get("continue_editing") != "" || ctx.Request.PostForm.Get("QorUpdateState") != "" {
					http.Redirect(ctx.Writer, ctx.Request, url, http.StatusFound)
					return
				}
				httpu.Redirect(ctx.Writer, ctx.Request, url, http.StatusFound)
			}).With([]string{"json", "xml"}, func() {
				if ctx.Request.FormValue("qorInlineEdit") != "" {
					rresult := reflect.ValueOf(record)
					for rresult.Kind() == reflect.Ptr {
						rresult = rresult.Elem()
					}
					newResult := make(map[string]interface{})

					for key := range ctx.Request.Form {
						if strings.HasPrefix(key, "QorResource.") {
							key = strings.TrimPrefix(key, "QorResource.")
							f := rresult.FieldByName(key)
							if f.IsValid() {
								newResult[key] = f.Interface()
							} else if v, ok := field_access.Get(record, key); ok {
								newResult[key] = v
							}
						}
					}
					newResult = valuesmap.ParseMap(newResult)
					ctx.Encode(newResult)
					return
				}
				ctx.Encode(record)
			}).XAccept().Respond(ctx.Request)
		}
	)

	if !ctx.HasError() {
		ctx.Result = record
		ctx.ResourceRecord = record

		if cfg.Prepare != nil {
			cfg.Prepare(ctx)
			if ctx.Writer.WroteHeader() {
				return
			}
		}

		old = res.New()

		if err := copier.Copy(old, record); err != nil {
			ctx.AddError(err)
			goto done
		}

		excludes := &core.DecoderExcludes{}
		ctx.Context.DecoderExcludes = excludes

		decerr := res.NewDecoder(ctx.Context, record).
			SetNotLoad(true).
			Decode(record)

		if ctx.AddError(decerr); !ctx.HasError() {
			if !notLoad && old == nil {
				if reader, ok := this.controller.(ControllerReader); ok {
					ctx.WithDB(func(context *Context) {
						context.AddError(context.CloneErr(func(ctx *Context) {
							old = reader.Read(ctx)
						}))
					}, ctx.DB().Set(aorm.OptKeySkipPreload, true))
				}
			}

			if stateName := ctx.Request.PostForm.Get("QorUpdateState"); stateName != "" {
				for _, state := range ctx.Resource.UpdateStates {
					if state.Name == stateName {
						err := state.Handler(ctx, &messages, func() {
							ctx.Type |= RE_RENDER
							if len(messages) > 0 {
								ctx.FlashS("success", messages...)
							}
							beforeRender(ctx, cfg)
							ctx.Writer.Header().Set("X-Frame-Reload", "render-body")
							ctx.Execute("edit", ctx.ResourceRecord)
						})
						if err != nil {
							ctx.AddError(err)
							goto done
						}
						return
					}
				}

				ctx.AddError(fmt.Errorf("State %q does not exists", stateName))
			}

			if !ctx.HasError() {
				this.controller.(ControllerUpdater).Update(ctx, record, old)
			}
		}
	}

done:
	ctx.Result = record

	beforeRender(ctx, cfg)

	if ctx.HasError() {
		var validationErrors = ctx.Errors.Filter(func(err error) error {
			if cause := validations.Cause(err); cause != nil {
				if aorm.IsQueryError(err) {
					return cause
				}
				return err
			}
			return nil
		})
		ctx.Writer.WriteHeader(HTTPUnprocessableEntity)
		responder.With("html", func() {
			if ctx.Errors.Len() != validationErrors.Len() {
				if cfg.ErrorCallback != nil {
					cfg.ErrorCallback(ctx, ctx.Errors)
					if ctx.Writer.WroteHeader() {
						return
					}
				}

				ctx.Execute("shared/errors", record)
			} else {
				ctx.Errors = validationErrors
				ctx.Execute("", record)
			}
		}).With([]string{"json", "xml"}, func() {
			ctx.Encode(map[string]interface{}{"errors": ctx.GetErrors()})
		}).Respond(ctx.Request)
	} else {
		if ctx.Resource.Config.Wizard != nil {
			wz := record.(WizardModelInterface)
			if wz.IsDone() {
				var (
					flashMessageDisabled bool
					url                  = cfg.RedirectTo
					status               = http.StatusSeeOther
					done                 bool
				)

				if cfg := ctx.WizardCompleteConfig; cfg != nil {
					flashMessageDisabled = cfg.FlashMessageDisabled
					if cfg.RedirectTo != "" {
						url = cfg.RedirectTo
					}
					if cfg.RedirectToStatus != 0 {
						status = cfg.RedirectToStatus
					}

					if cfg.RedirectToOther {
						ctx.Writer.Header().Set("X-Next-Step", "main")
					} else {
						ctx.Writer.Header().Set("X-Steps-Done", "true")
						done = true
					}
				} else {
					ctx.Writer.Header().Set("X-Steps-Done", "true")
					done = true
				}

				destRes := ctx.Resource.ParentResource
				dest := wz.GetDestination()

				if !flashMessageDisabled {
					message := string(ctx.tt(I18NGROUP+".form.successfully_created",
						NewResourceRecorde(ctx, destRes, dest), "{{.}} was successfully created"))
					if cfg.SuccessCallback != nil {
						cfg.SuccessCallback(ctx, old, record, &message)
						if ctx.Writer.WroteHeader() {
							return
						}
					}
					ctx.Flash(message, "success")
				}

				if url == "" {
					if ctx.Request.URL.Query().Get("continue_editing") != "" || ctx.Request.URL.Query().Get("continue_editing_url") != "" {
						url = res.ParentResource.GetContextURI(ctx, destRes.GetKey(dest)) + P_OBJ_UPDATE_FORM
					} else {
						url = res.ParentResource.GetContextIndexURI(ctx)
					}
				}

				if done && httpu.IsActionFormRequest(ctx.Request) {
					ctx.Writer.WriteHeader(http.StatusOK)
				} else {
					httpu.Redirect(ctx.Writer, ctx.Request, url, status)
				}
				return
			} else {
				ctx.Writer.Header().Set("X-Next-Step", wz.CurrentStepName())
				if httpu.IsXhrRequest(ctx.Request) {
					ctx.Writer.Header().Set("X-Location", ctx.Request.RequestURI)
					ctx.Writer.WriteHeader(http.StatusOK)
					return
				}
				httpu.Redirect(ctx.Writer, ctx.Request, ctx.Request.RequestURI, http.StatusSeeOther)
				return
			}
		}

		if !messageDisabled {
			message := string(ctx.tt(I18NGROUP+".form.successfully_updated", NewResourceRecorde(ctx, res, record),
				"{{.}} was successfully updated"))
			if cfg.SuccessCallback != nil {
				cfg.SuccessCallback(ctx, old, record, &message)
				if ctx.Writer.WroteHeader() {
					return
				}
			}
			messages = append(messages, message)
		}
		respond(messages)
	}
}
