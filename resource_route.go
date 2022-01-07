package admin

import (
	"github.com/ecletus/core"
	"github.com/ecletus/core/resource"
	"github.com/ecletus/core/utils"

	"github.com/moisespsena-go/xroute"
)

func (this *Resource) RegisterDefaultRouters(actions ...string) {
	this.ControllerBuilder.AppendDefaultActions(actions...)
	this.ViewControllerBuilder.RegisterDefaultHandlers()
	this.ControllerBuilder.RegisterDefaultRouters()
}

func (this *Resource) InitRoutes(parent *RouteNode) *xroute.Mux {
	// force prepare all metas
	this.SetupMetas()

	node := &this.Routes

	node.Handler = this.Router
	node.Menu = this.defaultMenu
	node.Resource = this

	node.Walk(func(_ []*RouteNodeWalkerItem, item *RouteNodeWalkerItem) error {
		for method, handler := range item.Node.MethodHandlers {
			this.Router.HandleMethod(method, item.Path, handler)
		}
		return nil
	})

	parent.Add(this.Param, node)

	if this.Config.Singleton {
		for param, subRes := range this.ResourcesByParam {
			r := subRes.InitRoutes(node)
			pattern := "/" + param
			this.Router.Mount(pattern, r)
		}
	} else {
		itemNode := node.addChild(this.ParamIDPattern(), &this.ItemRoutes)

		for param, subRes := range this.ResourcesByParam {
			if subRes.Config.Sub == nil || !subRes.Config.Sub.MountAsItemDisabled {
				r := subRes.InitRoutes(itemNode)
				pattern := "/" + param
				this.ItemRouter.Mount(pattern, r)
			}
		}

		for param, subRes := range this.ResourcesByParam {
			if subRes.Config.Sub != nil && subRes.Config.Sub.MountAsItemDisabled {
				r := subRes.InitRoutes(itemNode)
				pattern := "/" + param
				this.Router.Mount(pattern, r)
			}
		}

		this.ItemRoutes.Resource = this
		this.ItemRoutes.ResourceItem = true
		this.ItemRoutes.Handler = this.ItemRouter
		this.ItemRoutes.Walk(func(_ []*RouteNodeWalkerItem, item *RouteNodeWalkerItem) error {
			for method, handler := range item.Node.MethodHandlers {
				this.ItemRouter.HandleMethod(method, item.Path, handler)
			}
			return nil
		})

		this.Router.Mount("/"+this.ParamIDPattern(), this.ItemRouter)
	}

	for name, s := range this.Scheme.Children {
		node.Add(name, &RouteNode{Scheme: s})
	}

	for _, action := range this.Actions {
		if action.Resource != nil {
			action.Resource.SetupMetas()
		}
	}

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
	if this.mounted {
		if this.ControllerBuilder.IsIndexer() {
			this.setupMetas(this.IndexAttrs().ToMetas()...)
		}
		if this.ControllerBuilder.IsCreator() {
			this.setupMetas(this.NewAttrs().ToMetas()...)
		}
		if this.ControllerBuilder.IsUpdater() {
			this.setupMetas(this.EditAttrs().ToMetas()...)
		}
		if this.ControllerBuilder.IsReader() {
			this.setupMetas(this.ShowAttrs().ToMetas()...)
		}
		for _, l := range this.Sections.Layouts.Layouts {
			if l != this.Sections.Default {
				l.Print.MetasNamesCb(func(name string) {
					meta := this.GetMeta(name)
					if meta != nil && meta.Resource != nil {
						meta.Resource.SetupMetas()
					}
				})
			}
		}
		if !this.Config.Singleton {
			if _, ok := this.MetasByName["id"]; !ok {
				this.Meta(&Meta{
					Name: "id",
					Valuer: func(record interface{}, context *core.Context) interface{} {
						return this.GetKey(record)
					},
					Type: "hidden_primary_key",
					Setter: func(record interface{}, metaValue *resource.MetaValue, context *core.Context) error {
						if v := metaValue.FirstStringValue(); v != "" {
							if ID, err := this.ParseID(v); err != nil {
								return err
							} else {
								ID.SetTo(record)
							}
						}
						return nil
					},
				})
			}
		}
	} else {
		this.AllSections()
	}

	for _, f := range this.postMetasSetupCallbacks {
		f()
	}

	for _, action := range this.Actions {
		if action.Resource != nil {
			action.Resource.AllSectionsFunc()
		}
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
