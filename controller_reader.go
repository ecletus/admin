package admin

import (
	"net/http"

	"github.com/ecletus/roles"
	"github.com/moisespsena/template/html/template"

	"github.com/ecletus/core"
)

func ParseShowConfig(context *Context) (cfg *ShowConfig) {
	if tcfg := context.RouteHandler.Config.HandlerConfig; tcfg != nil {
		if opt, ok := tcfg.(*ShowConfig); ok {
			cfg2 := *opt
			cfg = &cfg2
		}
	}

	if cfg == nil {
		cfg = &ShowConfig{}
		cfg.Load(core.Configors{context.RouteHandler.Config, context.Resource}, context)
	}

	if cfg.ActionsDisabled {
		context.Options(OptContextActionsDisabled())
	}

	context.Funcs(template.FuncMap{
		"show_config": func() *ShowConfig {
			return cfg
		},
	})

	return
}

// Show render show page
func (this *Controller) Show(ctx *Context) {
	if _, ok := this.controller.(ControllerReader); !ok {
		ctx.NotFound = true
		http.NotFound(ctx.Writer, ctx.Request)
	}

	if ctx = ctx.ParentPreload(ParentPreloadShow); ctx.HasError() {
		ctx.LogErrors()
		return
	}

	ctx.SetBasicType(SHOW)
	if HasDeletedUrlQuery(ctx.Request.URL.Query()) {
		ctx.Type |= DELETED
	}
	this.showOrEdit(ctx, func(record interface{}) bool {
		ctx.Result = record
		ctx.ResourceRecord = record

		if ctx.Type.Has(EDIT) {
			ParseUpdateConfig(ctx)
		} else {
			ParseShowConfig(ctx)
		}
		return true
	})
}

func (this *Controller) LoadShowData(context *Context) (result interface{}, notExists bool) {
	if setuper, ok := this.controller.(ControllerSetuper); ok {
		if context.AddError(setuper.SetupContext(context)); context.HasError() {
			return
		}
	}

	res := context.Resource

	if res.Config.Singleton {
		if reader, ok := this.controller.(ControllerReader); ok {
			result = reader.Read(context)
		} else if creator, ok := this.controller.(ControllerCreator); ok {
			result = creator.New(context)
		} else {
			result = res.NewStruct(context.Site)
		}
		if result == nil {
			notExists = true
			if context.HasPermission(res, roles.Create) {
				if creator, ok := this.controller.(ControllerCreator); ok {
					result = creator.New(context)
				} else {
					result = res.NewStruct(context.Site)
				}
			}
		}
	} else {
		context.SetRawDB(context.DB().Unscoped())
		result = this.controller.(ControllerReader).Read(context)
	}

	if result != nil {
		for _, f := range GetOptContextRecordLoaded(context, context.RouteHandler.Config, context.Resource) {
			f(context, result)
		}
	}

	return
}
