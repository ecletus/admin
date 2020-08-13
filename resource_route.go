package admin

import (
	"github.com/ecletus/core/utils"

	"github.com/moisespsena-go/xroute"
)

func (this *Resource) RegisterDefaultRouters(actions ...string) {
	this.ControllerBuilder.AppendDefaultActions(actions...)
	this.ViewControllerBuilder.RegisterDefaultHandlers()
	this.ControllerBuilder.RegisterDefaultRouters()
}

func (this *Resource) InitRoutes() *xroute.Mux {
	if this.Config.Singleton {
		for param, subRes := range this.ResourcesByParam {
			r := subRes.InitRoutes()
			pattern := "/" + param
			this.Router.Mount(pattern, r)
		}
	} else {
		for param, subRes := range this.ResourcesByParam {
			r := subRes.InitRoutes()
			pattern := "/" + param
			this.ItemRouter.Mount(pattern, r)
		}
		this.Router.Mount("/"+this.ParamIDPattern(), this.ItemRouter)
	}

	// force prepare all metas
	this.SetupMetas()
	return this.Router
}

func (this *Resource) setupMetas(metas ...*Meta) {
	for _, meta := range metas {
		if meta.Resource != nil {
			meta.Resource.SetupMetas()
		}
	}
}

func (this *Resource) SetupMetas() {
	if this.setupMetasCalled {
		return
	}
	this.setupMetasCalled = true
	if this.ControllerBuilder.IsIndexer() {
		this.setupMetas(this.ConvertSectionToMetas(this.IndexAttrs())...)
	}
	if this.ControllerBuilder.IsCreator() {
		this.setupMetas(this.ConvertSectionToMetas(this.NewAttrs())...)
	}
	if this.ControllerBuilder.IsUpdater() {
		this.setupMetas(this.ConvertSectionToMetas(this.EditAttrs())...)
	}
	if this.ControllerBuilder.IsReader() {
		this.setupMetas(this.ConvertSectionToMetas(this.ShowAttrs())...)
	}
}

func (this *Resource) MountTo(param string) *Resource {
	config := &(*this.Config)
	if config.Sub != nil {
		config.Sub = &(*config.Sub)
	}
	nmp := utils.NamifyString(param)
	config.Name += nmp
	config.Param = param
	config.ID += nmp
	config.NotMount = false
	config.Invisible = true
	return this.Admin.AddResource(this.Value, config)
}